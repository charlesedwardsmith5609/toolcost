package scripts

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"
)

// fixture is a throwaway git repo with the review scripts, a fake OpenAI endpoint on 127.0.0.1,
// and a PATH that holds only what the scripts need plus an optional fake gitleaks.
type fixture struct {
	t      *testing.T
	repo   string
	bin    string
	mu     sync.Mutex
	bodies []string
}

func copyFile(t *testing.T, src, dst string) {
	t.Helper()
	data, err := os.ReadFile(src)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dst, data, 0o644); err != nil {
		t.Fatal(err)
	}
}

func write(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func (f *fixture) git(args ...string) string {
	f.t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = f.repo
	cmd.Env = append(os.Environ(), "GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@example.com",
		"GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@example.com")
	out, err := cmd.CombinedOutput()
	if err != nil {
		f.t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return string(out)
}

func newFixture(t *testing.T) *fixture {
	t.Helper()
	for _, tool := range []string{"git", "bash", "python3", "curl"} {
		if _, err := exec.LookPath(tool); err != nil {
			t.Skip(tool, "not available")
		}
	}
	f := &fixture{t: t, repo: t.TempDir(), bin: t.TempDir()}
	for _, name := range []string{"cross-review.sh", "review-call.sh"} {
		copyFile(t, name, filepath.Join(f.repo, "scripts", name))
	}
	copyFile(t, filepath.Join("..", "skills", "review-checklist.md"), filepath.Join(f.repo, "skills", "review-checklist.md"))
	f.git("init", "-q")
	write(t, filepath.Join(f.repo, "app.txt"), "one\n")
	write(t, filepath.Join(f.repo, "server.key"), "key v1\n")
	write(t, filepath.Join(f.repo, ".gitignore"), ".reviews/\n")
	f.git("add", "app.txt", "server.key", ".gitignore", "scripts", "skills")
	f.git("commit", "-q", "-m", "init")
	write(t, filepath.Join(f.repo, "app.txt"), "one\nTRACKED_CHANGE_MARKER\n")
	write(t, filepath.Join(f.repo, "server.key"), "key v2 TRACKED_KEY_MARKER\n")

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		f.mu.Lock()
		f.bodies = append(f.bodies, string(body))
		f.mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"looks fine"}}]}`))
	}))
	t.Cleanup(srv.Close)
	t.Setenv("OPENAI_BASE_URL", srv.URL)
	return f
}

func (f *fixture) fakeGitleaks(exit int) {
	script := "#!/bin/sh\nexit " + string(rune('0'+exit)) + "\n"
	path := filepath.Join(f.bin, "gitleaks")
	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		f.t.Fatal(err)
	}
}

// run executes cross-review.sh with PATH limited to the fake bin dir and the system dirs.
func (f *fixture) run(args ...string) (string, int) {
	f.t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "bash", append([]string{"scripts/cross-review.sh", "--provider", "openai"}, args...)...)
	cmd.Dir = f.repo
	cmd.Env = []string{
		"PATH=" + f.bin + ":/usr/bin:/bin",
		"HOME=" + f.t.TempDir(),
		"PAIRING_KIT_ENV=" + filepath.Join(f.t.TempDir(), "none.env"),
		"OPENAI_API_KEY=test-key",
		"OPENAI_BASE_URL=" + os.Getenv("OPENAI_BASE_URL"),
	}
	out, err := cmd.CombinedOutput()
	if ctx.Err() != nil {
		f.t.Fatalf("timed out\n%s", out)
	}
	if err == nil {
		return string(out), 0
	}
	ee, ok := err.(*exec.ExitError)
	if !ok {
		f.t.Fatal(err)
	}
	return string(out), ee.ExitCode()
}

func (f *fixture) sent() string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return strings.Join(f.bodies, "\n")
}

func TestNothingSentWithoutSecretScan(t *testing.T) {
	f := newFixture(t) // no gitleaks on PATH
	if _, err := exec.LookPath("gitleaks"); err == nil {
		// The fixture PATH is /usr/bin:/bin only; a system-wide gitleaks there would defeat this case.
		if p, _ := exec.LookPath("gitleaks"); strings.HasPrefix(p, "/usr/bin") || strings.HasPrefix(p, "/bin") {
			t.Skip("gitleaks is installed in a system directory")
		}
	}
	out, code := f.run()
	if code == 0 || f.sent() != "" {
		t.Fatalf("missing gitleaks: exit %d, sent %q\n%s", code, f.sent(), out)
	}
	if !strings.Contains(out, "Nothing was sent") {
		t.Fatalf("no explanation:\n%s", out)
	}
}

func TestNothingSentWhenScanFailsOrFinds(t *testing.T) {
	for _, exit := range []int{7, 3} {
		f := newFixture(t)
		f.fakeGitleaks(exit)
		out, code := f.run()
		if code == 0 || f.sent() != "" {
			t.Fatalf("gitleaks exit %d: exit %d, sent %q\n%s", exit, code, f.sent(), out)
		}
	}
}

func TestDefaultSendsTrackedOnlyAndSkipsSecretFiles(t *testing.T) {
	f := newFixture(t)
	f.fakeGitleaks(0)
	write(t, filepath.Join(f.repo, ".env"), "API_TOKEN=UNTRACKED_ENV_MARKER\n")
	write(t, filepath.Join(f.repo, "notes.txt"), "UNTRACKED_NOTES_MARKER\n")
	out, code := f.run()
	if code != 0 {
		t.Fatalf("exit %d\n%s", code, out)
	}
	sent := f.sent()
	if !strings.Contains(sent, "TRACKED_CHANGE_MARKER") {
		t.Fatalf("tracked change missing from request: %s", sent)
	}
	for _, bad := range []string{"UNTRACKED_ENV_MARKER", "UNTRACKED_NOTES_MARKER", "TRACKED_KEY_MARKER"} {
		if strings.Contains(sent, bad) {
			t.Errorf("%s was sent", bad)
		}
	}
	if status := f.git("status", "--porcelain"); !strings.Contains(status, "?? notes.txt") {
		t.Errorf("the script changed the index; status:\n%s", status)
	}
}

func TestIncludeNamesUntrackedFile(t *testing.T) {
	f := newFixture(t)
	f.fakeGitleaks(0)
	write(t, filepath.Join(f.repo, ".env"), "API_TOKEN=UNTRACKED_ENV_MARKER\n")
	write(t, filepath.Join(f.repo, "notes.txt"), "UNTRACKED_NOTES_MARKER\n")
	if out, code := f.run("--include", "notes.txt"); code != 0 || !strings.Contains(f.sent(), "UNTRACKED_NOTES_MARKER") {
		t.Fatalf("exit %d, sent %q\n%s", code, f.sent(), out)
	}
	if strings.Contains(f.sent(), "UNTRACKED_ENV_MARKER") {
		t.Fatal(".env was sent")
	}
}

func TestIncludeRefusesSecretPathsAndSymlinks(t *testing.T) {
	f := newFixture(t)
	f.fakeGitleaks(0)
	write(t, filepath.Join(f.repo, ".env"), "UNTRACKED_ENV_MARKER\n")
	write(t, filepath.Join(f.repo, ".env.local"), "UNTRACKED_ENV_MARKER\n")
	write(t, filepath.Join(f.repo, "deploy", "id_ed25519"), "UNTRACKED_ENV_MARKER\n")
	write(t, filepath.Join(f.repo, "aws-credentials.txt"), "UNTRACKED_ENV_MARKER\n")
	write(t, filepath.Join(f.repo, ".aws", "config"), "UNTRACKED_ENV_MARKER\n")
	write(t, filepath.Join(f.repo, "notes.txt"), "ok\n")
	if err := os.Symlink("notes.txt", filepath.Join(f.repo, "link.txt")); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{".env", "./.env.local", "deploy/id_ed25519", "aws-credentials.txt", ".aws/config", "server.key", "link.txt", "../outside.txt"} {
		out, code := f.run("--include", path)
		if code != 2 || f.sent() != "" {
			t.Errorf("--include %s: exit %d, sent %q\n%s", path, code, f.sent(), out)
		}
	}
}

// The .sh and .ps1 deny lists must stay identical.
func TestDenyListsMatch(t *testing.T) {
	list := func(file, name string) string {
		data, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		re := regexp.MustCompile(`(?s)\$?` + name + `=?\s*=?\s*[@(]+?\(?(.*?)\)`)
		m := re.FindSubmatch(data)
		if m == nil {
			t.Fatalf("%s: %s not found", file, name)
		}
		return regexp.MustCompile(`[\s,]+`).ReplaceAllString(strings.TrimSpace(string(m[1])), " ")
	}
	for _, name := range []struct{ sh, ps string }{{"deny_names", "denyNames"}, {"deny_dirs", "denyDirs"}} {
		if a, b := list("cross-review.sh", name.sh), list("cross-review.ps1", name.ps); a != b {
			t.Errorf("deny lists differ:\n sh: %s\n ps: %s", a, b)
		}
	}
}
