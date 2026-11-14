package estimate

import (
	"testing"

	"toolcost/internal/load"
)

type fixed struct{ n int }

func (f fixed) Tokens([]byte) int { return f.n }

func TestStubEstimatorAndBudgetMath(t *testing.T) {
	items := []load.Item{{Name: "big", Kind: "tool", Schema: []byte("xxxx")}, {Name: "small", Kind: "mcp", Schema: []byte("x")}}
	report := Count(items, fixed{n: 10})
	if report.Total != 20 || report.Estimator != "custom" {
		t.Fatalf("%+v", report)
	}
	if report.Rows[0].Name != "big" && report.Rows[1].Name != "big" {
		t.Fatal(report.Rows)
	}
	bytes := BytesPerToken{Divisor: 4}
	if bytes.Tokens([]byte("12345")) != 2 || bytes.Tokens(nil) != 0 {
		t.Fatal(bytes.Tokens([]byte("12345")))
	}
	wide := Count([]load.Item{{Name: "only", Schema: bytesOf(800)}}, Default())
	if wide.Total < 200 || len(wide.Suggestions) != 1 {
		t.Fatalf("%+v", wide)
	}
}

func bytesOf(n int) []byte {
	out := make([]byte, n)
	for i := range out {
		out[i] = 'a'
	}
	return out
}
