package report

import (
	"bytes"
	"strings"
	"testing"

	"toolcost/internal/estimate"
)

func TestTextNamesTheTotal(t *testing.T) {
	var buf bytes.Buffer
	err := Text(&buf, estimate.Report{Total: 12, Estimator: "bytes/4", Rows: []estimate.Row{{Name: "search", Kind: "tool", Tokens: 12}}})
	if err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "total 12 tokens per turn") || strings.Contains(out, "0.5x") {
		t.Fatal(out)
	}
	if !strings.Contains(out, "estimator bytes/4\n") {
		t.Fatal(out)
	}
}

func TestTextEscapesNames(t *testing.T) {
	var buf bytes.Buffer
	evil := "evil\x1b[31m\nforged: line"
	rep := estimate.Report{
		Total:       5,
		Estimator:   "bytes/4",
		Rows:        []estimate.Row{{Name: evil, Kind: "tool", Tokens: 5, Suggest: "Disable " + evil}},
		Suggestions: []string{"Disable " + evil},
	}
	if err := Text(&buf, rep); err != nil {
		t.Fatal(err)
	}
	for _, r := range buf.String() {
		if r < 0x20 && r != '\n' || r == 0x7f {
			t.Fatalf("raw control character %q in %q", r, buf.String())
		}
	}
	if strings.Contains(buf.String(), "\nforged: line") {
		t.Fatalf("name forged a line: %q", buf.String())
	}
}

func TestTextLabelsMCPConfigOnly(t *testing.T) {
	var buf bytes.Buffer
	rep := estimate.Report{Total: 0, MCPConfigTokens: 9, Estimator: "bytes/4",
		Rows: []estimate.Row{{Name: "fs", Kind: "mcp", Tokens: 9, Note: "config only"}}}
	if err := Text(&buf, rep); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "config only") || !strings.Contains(out, "mcp config 9 tokens") {
		t.Fatal(out)
	}
}
