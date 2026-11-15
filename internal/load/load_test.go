package load

import (
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
)

func TestParseSortsKeys(t *testing.T) {
	items, err := Parse("tools.json", []byte(`{"mcpServers":{"b":{"command":"b"},"a":{"command":"a"}},"tools":[{"name":"search","input_schema":{"z":1,"a":2}}]}`))
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 3 {
		t.Fatalf("%d items", len(items))
	}
	if items[0].Name != "search" || string(items[0].Schema) != `{"input_schema":{"a":2,"z":1},"name":"search"}` {
		t.Fatalf("tool: %+v %s", items[0], items[0].Schema)
	}
	if items[1].Name != "a" || items[2].Name != "b" {
		t.Fatalf("server order: %+v", items)
	}
}

// readWithTimeout runs ReadFile in a goroutine so a hang fails the test.
func readWithTimeout(t *testing.T, path string) ([]Item, error) {
	t.Helper()
	type result struct {
		items []Item
		err   error
	}
	done := make(chan result, 1)
	go func() {
		items, err := ReadFile(path)
		done <- result{items, err}
	}()
	select {
	case r := <-done:
		return r.items, r.err
	case <-time.After(5 * time.Second):
		t.Fatalf("ReadFile(%s) did not return within 5s", path)
		return nil, nil
	}
}

func TestReadFileRefusesSymlink(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "t.json")
	if err := os.WriteFile(target, []byte(`[]`), 0o644); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(dir, "link.json")
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}
	if _, err := readWithTimeout(t, link); err == nil {
		t.Fatal("symlink should be refused")
	}
}

func TestReadFileRefusesFIFO(t *testing.T) {
	path := filepath.Join(t.TempDir(), "fifo.json")
	if err := syscall.Mkfifo(path, 0o600); err != nil {
		t.Skip("mkfifo unavailable:", err)
	}
	if _, err := readWithTimeout(t, path); err == nil {
		t.Fatal("FIFO should be refused")
	}
}

func TestReadFileRefusesDevZero(t *testing.T) {
	if _, err := os.Stat("/dev/zero"); err != nil {
		t.Skip("no /dev/zero")
	}
	_, err := readWithTimeout(t, "/dev/zero")
	if err == nil || !strings.Contains(err.Error(), "regular") {
		t.Fatalf("want a not-regular error, got %v", err)
	}
}

func TestReadFileSizeLimit(t *testing.T) {
	dir := t.TempDir()
	// A bare array padded with spaces to an exact size.
	pad := func(n int) []byte {
		return []byte("[]" + strings.Repeat(" ", n-2))
	}
	exact := filepath.Join(dir, "exact.json")
	if err := os.WriteFile(exact, pad(MaxBytes), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := readWithTimeout(t, exact); err != nil {
		t.Fatalf("exactly MaxBytes should be accepted: %v", err)
	}
	over := filepath.Join(dir, "over.json")
	if err := os.WriteFile(over, pad(MaxBytes+1), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := readWithTimeout(t, over); err == nil {
		t.Fatal("MaxBytes+1 should be refused")
	}
}

func TestParseShapes(t *testing.T) {
	bad := map[string]string{
		"empty object":       `{}`,
		"unknown key":        `{"foo":1}`,
		"null":               `null`,
		"tools null":         `{"tools":null}`,
		"tools object":       `{"tools":{}}`,
		"mcpServers array":   `{"mcpServers":[]}`,
		"unknown plus known": `{"tools":[],"toolz":[]}`,
		"scalar":             `7`,
		"two arrays":         `[] []`,
		"trailing garbage":   `{"tools":[]} x`,
		"two objects":        `{"tools":[]}{"tools":[]}`,
		"duplicate key":      `{"tools":[],"tools":[]}`,
		"nested duplicate":   `{"tools":[{"name":"a","name":"b"}]}`,
		"too deep":           `[` + strings.Repeat(`[`, 64) + strings.Repeat(`]`, 64) + `]`,
	}
	for name, in := range bad {
		if _, err := Parse("x.json", []byte(in)); err == nil {
			t.Errorf("%s: %s should be an error", name, in)
		}
	}
	good := map[string]int{
		`{"tools":[]}`: 0,
		`[]`:           0,
		`{"tools":[{"name":"a"}],"mcpServers":{"s":{"command":"c"}}}`: 2,
		`{"mcpServers":{}}`: 0,
		`[` + strings.Repeat(`[`, 62) + strings.Repeat(`]`, 62) + `]`: 1,
	}
	for in, n := range good {
		items, err := Parse("x.json", []byte(in))
		if err != nil {
			t.Errorf("%s: %v", in, err)
			continue
		}
		if len(items) != n {
			t.Errorf("%s: %d items, want %d", in, len(items), n)
		}
	}
}
