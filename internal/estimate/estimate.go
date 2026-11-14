// Package estimate counts approximate tokens for tool schemas.
package estimate

import (
	"errors"
	"fmt"
	"math"
	"sort"

	"toolcost/internal/load"
)

// Estimator counts tokens for a canonical schema. Implementations must not use the network.
type Estimator interface {
	Tokens(schema []byte) int
}

// MaxDivisor is the largest divisor Validate accepts.
const MaxDivisor = 1 << 20

// BytesPerToken estimates tokens as len(schema)/Divisor, rounded up.
// Divisor 4 is the default. The result is an estimate; it is not measured against a vendor tokenizer.
type BytesPerToken struct {
	Divisor int
}

// Validate rejects a divisor outside 1..MaxDivisor.
func (e BytesPerToken) Validate() error {
	if e.Divisor < 1 || e.Divisor > MaxDivisor {
		return fmt.Errorf("divisor %d is out of range 1..%d", e.Divisor, MaxDivisor)
	}
	return nil
}

// Name implements Named.
func (e BytesPerToken) Name() string { return fmt.Sprintf("bytes/%d", e.Divisor) }

// Tokens implements Estimator. Count validates the divisor first; a bad divisor yields 0 here
// instead of being replaced by another value.
func (e BytesPerToken) Tokens(schema []byte) int {
	if e.Validate() != nil || len(schema) == 0 {
		return 0
	}
	return (len(schema) + e.Divisor - 1) / e.Divisor
}

// Named is implemented by estimators that can describe themselves. Others report "custom".
type Named interface {
	Name() string
}

// validator is implemented by estimators with parameters to check.
type validator interface {
	Validate() error
}

// Default is the documented bytes/4 estimator.
func Default() Estimator { return BytesPerToken{Divisor: 4} }

// Row is one counted tool or server. MCP rows measure launch config only.
type Row struct {
	Name    string `json:"name"`
	Kind    string `json:"kind"`
	Tokens  int    `json:"tokens"`
	Note    string `json:"note,omitempty"`
	Suggest string `json:"suggest,omitempty"`
}

// Report is the per-turn overhead. Total and the budget count tool rows only; MCP server rows
// are config only and are summed separately in MCPConfigTokens.
type Report struct {
	Total           int      `json:"total"`
	MCPConfigTokens int      `json:"mcp_config_tokens"`
	Budget          *int     `json:"budget,omitempty"`
	Estimator       string   `json:"estimator"`
	Rows            []Row    `json:"rows"`
	Suggestions     []string `json:"suggestions"`
}

const (
	minSuggestTotal  = 200
	minSuggestTokens = 200
)

func addChecked(a, b int) (int, error) {
	if b < 0 || a > math.MaxInt-b {
		return 0, errors.New("token count overflows")
	}
	return a + b, nil
}

// Count applies est to each item. It returns an error for an invalid estimator, a negative
// count, or a sum that would overflow.
func Count(items []load.Item, est Estimator) (Report, error) {
	if v, ok := est.(validator); ok {
		if err := v.Validate(); err != nil {
			return Report{}, err
		}
	}
	report := Report{Estimator: "custom", Rows: []Row{}}
	if n, ok := est.(Named); ok {
		report.Estimator = n.Name()
	}
	for _, item := range items {
		tokens := est.Tokens(item.Schema)
		if tokens < 0 {
			return Report{}, errors.New("estimator returned a negative count")
		}
		row := Row{Name: item.Name, Kind: item.Kind, Tokens: tokens}
		var err error
		if item.Kind == "mcp" {
			row.Note = "config only"
			report.MCPConfigTokens, err = addChecked(report.MCPConfigTokens, tokens)
		} else {
			report.Total, err = addChecked(report.Total, tokens)
		}
		if err != nil {
			return Report{}, err
		}
		report.Rows = append(report.Rows, row)
	}
	sort.Slice(report.Rows, func(i, j int) bool {
		if report.Rows[i].Tokens == report.Rows[j].Tokens {
			return report.Rows[i].Name < report.Rows[j].Name
		}
		return report.Rows[i].Tokens > report.Rows[j].Tokens
	})
	// Suggestions are for small-setup noise control: none below minSuggestTotal, and the
	// share rule needs at least three tool rows.
	toolRows := 0
	for _, row := range report.Rows {
		if row.Note == "" {
			toolRows++
		}
	}
	for i := range report.Rows {
		row := &report.Rows[i]
		if row.Note != "" || report.Total < minSuggestTotal {
			continue
		}
		if row.Tokens >= minSuggestTokens || (toolRows >= 3 && row.Tokens*100/report.Total >= 25) {
			row.Suggest = fmt.Sprintf("Disable %q on turns that do not need it.", row.Name)
			report.Suggestions = append(report.Suggestions, row.Suggest)
		}
	}
	if report.Suggestions == nil {
		report.Suggestions = []string{}
	}
	return report, nil
}
