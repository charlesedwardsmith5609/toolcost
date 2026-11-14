// Package estimate counts approximate tokens for tool schemas.
package estimate

import (
	"sort"

	"toolcost/internal/load"
)

// Estimator counts tokens for a canonical schema. Implementations must not use the network.
type Estimator interface {
	Tokens(schema []byte) int
}

// BytesPerToken estimates tokens as len(schema)/Divisor, rounded up.
// Divisor 4 is the default. For schema text this is about 0.5x to 2x a typical BPE count.
type BytesPerToken struct {
	Divisor int
}

// Tokens implements Estimator.
func (e BytesPerToken) Tokens(schema []byte) int {
	divisor := e.Divisor
	if divisor <= 0 {
		divisor = 4
	}
	if len(schema) == 0 {
		return 0
	}
	return (len(schema) + divisor - 1) / divisor
}

// Default is the documented bytes/4 estimator.
func Default() Estimator { return BytesPerToken{Divisor: 4} }

// Row is one counted tool or server.
type Row struct {
	Name    string `json:"name"`
	Kind    string `json:"kind"`
	Tokens  int    `json:"tokens"`
	Suggest string `json:"suggest,omitempty"`
}

// Report is the per-turn overhead.
type Report struct {
	Total       int      `json:"total"`
	Budget      int      `json:"budget,omitempty"`
	Estimator   string   `json:"estimator"`
	ErrorBound  string   `json:"error_bound"`
	Rows        []Row    `json:"rows"`
	Suggestions []string `json:"suggestions"`
}

// Count applies est to each item.
func Count(items []load.Item, est Estimator) Report {
	report := Report{
		Estimator:  "bytes/4",
		ErrorBound: "about 0.5x to 2x a typical BPE tokenizer on schema text",
		Rows:       []Row{},
	}
	if _, ok := est.(BytesPerToken); !ok {
		report.Estimator = "custom"
	}
	for _, item := range items {
		report.Rows = append(report.Rows, Row{Name: item.Name, Kind: item.Kind, Tokens: est.Tokens(item.Schema)})
		report.Total += report.Rows[len(report.Rows)-1].Tokens
	}
	sort.Slice(report.Rows, func(i, j int) bool {
		if report.Rows[i].Tokens == report.Rows[j].Tokens {
			return report.Rows[i].Name < report.Rows[j].Name
		}
		return report.Rows[i].Tokens > report.Rows[j].Tokens
	})
	for i := range report.Rows {
		row := &report.Rows[i]
		if report.Total > 0 && (row.Tokens >= 200 || row.Tokens*100/report.Total >= 25) {
			row.Suggest = "Disable " + row.Name + " on turns that do not need it."
			report.Suggestions = append(report.Suggestions, row.Suggest)
		}
	}
	if report.Suggestions == nil {
		report.Suggestions = []string{}
	}
	return report
}
