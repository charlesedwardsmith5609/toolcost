package main

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestBudget(t *testing.T) {
	bin := filepath.Join(t.TempDir(), "toolcost")
	build := exec.Command("go", "build", "-o", bin, ".")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build: %v\n%s", err, out)
	}
	root := filepath.Join("..", "..", "testdata")
	under := exec.Command(bin, "--format", "json", "--budget", "5000", filepath.Join(root, "small.json"))
	out, err := under.CombinedOutput()
	if err != nil {
		t.Fatalf("under budget: %v\n%s", err, out)
	}
	if !strings.Contains(string(out), `"estimator": "bytes/4"`) || strings.Contains(string(out), "input_schema") {
		t.Fatalf("json leaked a schema or lost the estimator: %s", out)
	}
	over := exec.Command(bin, "--budget", "1", filepath.Join(root, "large.json"))
	out, err = over.CombinedOutput()
	if err == nil {
		t.Fatal("over budget should exit non-zero")
	}
	if !strings.Contains(string(out), "suggest:") {
		t.Fatalf("text: %s", out)
	}
}
