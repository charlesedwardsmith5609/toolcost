package report

import (
	"bytes"
	"strings"
	"testing"

	"toolcost/internal/estimate"
)

func TestTextNamesTheTotal(t *testing.T) {
	var buf bytes.Buffer
	err := Text(&buf, estimate.Report{Total: 12, Estimator: "bytes/4", ErrorBound: "about 0.5x to 2x", Rows: []estimate.Row{{Name: "search", Kind: "tool", Tokens: 12}}})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "total 12 tokens per turn") {
		t.Fatal(buf.String())
	}
}
