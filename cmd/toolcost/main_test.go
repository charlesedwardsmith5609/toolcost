package main

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// buildBin compiles the CLI once per test.
func buildBin(t *testing.T) string {
	t.Helper()
	bin := filepath.Join(t.TempDir(), "toolcost")
	build := exec.Command("go", "build", "-o", bin, ".")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build: %v\n%s", err, out)
	}
	return bin
}

// run executes the binary and returns combined output and the exact exit code.
func run(t *testing.T, bin string, args ...string) (string, int) {
	t.Helper()
	out, err := exec.Command(bin, args...).CombinedOutput()
	if err == nil {
		return string(out), 0
	}
	var ee *exec.ExitError
	if !errors.As(err, &ee) {
		t.Fatalf("run: %v", err)
	}
	return string(out), ee.ExitCode()
}

func writeFile(t *testing.T, name, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestBudget(t *testing.T) {
	bin := buildBin(t)
	root := filepath.Join("..", "..", "testdata")
	out, code := run(t, bin, "--format", "json", "--budget", "5000", filepath.Join(root, "small.json"))
	if code != 0 {
		t.Fatalf("under budget: exit %d\n%s", code, out)
	}
	if !strings.Contains(out, `"estimator": "bytes/4"`) || strings.Contains(out, "input_schema") {
		t.Fatalf("json leaked a schema or lost the estimator: %s", out)
	}
	out, code = run(t, bin, "--budget", "1", filepath.Join(root, "large.json"))
	if code != 1 {
		t.Fatalf("over budget: exit %d, want 1\n%s", code, out)
	}
}

func TestBadInputExitsTwo(t *testing.T) {
	bin := buildBin(t)
	dir := t.TempDir()
	good := writeFile(t, "good.json", `{"tools":[]}`)
	link := filepath.Join(dir, "link.json")
	if err := os.Symlink(good, link); err != nil {
		t.Fatal(err)
	}
	cases := map[string]string{
		"empty object":  writeFile(t, "a.json", `{}`),
		"unknown key":   writeFile(t, "b.json", `{"foo":1}`),
		"null":          writeFile(t, "c.json", `null`),
		"trailing":      writeFile(t, "d.json", `{"tools":[]} {"tools":[]}`),
		"duplicate key": writeFile(t, "e.json", `{"tools":[],"tools":[]}`),
		"syntax":        writeFile(t, "f.json", `{`),
		"missing":       filepath.Join(dir, "nope.json"),
		"directory":     dir,
		"symlink":       link,
	}
	for name, path := range cases {
		out, code := run(t, bin, path)
		if code != 2 {
			t.Errorf("%s: exit %d, want 2\n%s", name, code, out)
		}
		out, code = run(t, bin, "--budget", "1000000", path)
		if code != 2 {
			t.Errorf("%s with budget: exit %d, want 2\n%s", name, code, out)
		}
	}
}
