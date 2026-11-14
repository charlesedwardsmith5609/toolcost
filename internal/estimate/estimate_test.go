package estimate

import (
	"fmt"
	"math"
	"strings"
	"testing"

	"toolcost/internal/load"
)

type fixed struct{ n int }

func (f fixed) Tokens([]byte) int { return f.n }

func TestStubEstimatorAndBudgetMath(t *testing.T) {
	items := []load.Item{{Name: "big", Kind: "tool", Schema: []byte("xxxx")}, {Name: "small", Kind: "tool", Schema: []byte("x")}}
	report, err := Count(items, fixed{n: 10})
	if err != nil {
		t.Fatal(err)
	}
	if report.Total != 20 || report.Estimator != "custom" {
		t.Fatalf("%+v", report)
	}
	bytes := BytesPerToken{Divisor: 4}
	if bytes.Tokens([]byte("12345")) != 2 || bytes.Tokens(nil) != 0 {
		t.Fatal(bytes.Tokens([]byte("12345")))
	}
	wide, err := Count([]load.Item{{Name: "only", Kind: "tool", Schema: bytesOf(800)}}, Default())
	if err != nil {
		t.Fatal(err)
	}
	if wide.Total < 200 || len(wide.Suggestions) != 1 {
		t.Fatalf("%+v", wide)
	}
}

func TestValidate(t *testing.T) {
	for _, d := range []int{0, -1, math.MinInt, 1 << 62, MaxDivisor + 1} {
		if err := (BytesPerToken{Divisor: d}).Validate(); err == nil {
			t.Errorf("divisor %d should be rejected", d)
		}
		if _, err := Count(nil, BytesPerToken{Divisor: d}); err == nil {
			t.Errorf("Count with divisor %d should fail", d)
		}
	}
	for _, d := range []int{1, 4, MaxDivisor} {
		if err := (BytesPerToken{Divisor: d}).Validate(); err != nil {
			t.Errorf("divisor %d: %v", d, err)
		}
	}
}

func TestSumDoesNotOverflow(t *testing.T) {
	items := []load.Item{{Name: "a", Kind: "tool", Schema: []byte("x")}, {Name: "b", Kind: "tool", Schema: []byte("x")}}
	if _, err := Count(items, fixed{n: math.MaxInt}); err == nil {
		t.Fatal("overflowing sum should be an error")
	}
	if _, err := Count(items, fixed{n: -1}); err == nil {
		t.Fatal("negative token count should be an error")
	}
}

func TestEstimatorName(t *testing.T) {
	cases := []struct {
		est  Estimator
		want string
	}{
		{BytesPerToken{Divisor: 4}, "bytes/4"},
		{BytesPerToken{Divisor: 8}, "bytes/8"},
		{fixed{n: 1}, "custom"},
	}
	for _, c := range cases {
		rep, err := Count(nil, c.est)
		if err != nil {
			t.Fatal(err)
		}
		if rep.Estimator != c.want {
			t.Errorf("got %q want %q", rep.Estimator, c.want)
		}
	}
}

func TestMCPRowsAreConfigOnly(t *testing.T) {
	items := []load.Item{
		{Name: "t", Kind: "tool", Schema: []byte("xxxx")},
		{Name: "s", Kind: "mcp", Schema: []byte("xxxxxxxx")},
	}
	rep, err := Count(items, Default())
	if err != nil {
		t.Fatal(err)
	}
	if rep.Total != 1 || rep.MCPConfigTokens != 2 {
		t.Fatalf("%+v", rep)
	}
	notes := map[string]string{}
	for _, r := range rep.Rows {
		notes[r.Name] = r.Note
	}
	if notes["s"] != "config only" || notes["t"] != "" {
		t.Fatalf("notes %v", notes)
	}
	only, err := Count(items[1:], Default())
	if err != nil {
		t.Fatal(err)
	}
	if only.Total != 0 || only.MCPConfigTokens != 2 {
		t.Fatalf("%+v", only)
	}
}

func TestNoErrorBoundField(t *testing.T) {
	rep, _ := Count(nil, Default())
	if strings.Contains(strings.ToLower(rep.Estimator), "0.5") {
		t.Fatal(rep.Estimator)
	}
}

func TestSuggestionRule(t *testing.T) {
	tools := func(sizes ...int) []load.Item {
		var out []load.Item
		for i, tokens := range sizes {
			out = append(out, load.Item{Name: fmt.Sprintf("t%d", i), Kind: "tool", Schema: []byte(strings.Repeat("x", tokens*4))})
		}
		return out
	}
	cases := []struct {
		name  string
		items []load.Item
		want  int
	}{
		{"one tool at 100 percent and 50 tokens", tools(50), 0},
		{"total 199 never", tools(199), 0},
		{"total 199 over many rows", tools(100, 99), 0},
		{"one row at 200 tokens", tools(200), 1},
		{"3+ rows, one at 30 percent, total 500", tools(150, 120, 120, 110), 1},
		{"two rows, share rule does not apply", tools(150, 50), 0},
		{"five rows at 20 percent", tools(100, 100, 100, 100, 100), 0},
	}
	for _, c := range cases {
		rep, err := Count(c.items, Default())
		if err != nil {
			t.Fatal(err)
		}
		if len(rep.Suggestions) != c.want {
			t.Errorf("%s: %d suggestions, want %d (%+v)", c.name, len(rep.Suggestions), c.want, rep)
		}
	}
}

func bytesOf(n int) []byte {
	out := make([]byte, n)
	for i := range out {
		out[i] = 'a'
	}
	return out
}
