package report

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"toolcost/internal/estimate"
)

// Text writes the per-turn total, the top offenders, and suggestions.
func Text(w io.Writer, rep estimate.Report) error {
	if _, err := fmt.Fprintf(w, "total %d tokens per turn\n", rep.Total); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "estimator %s\n", rep.Estimator); err != nil {
		return err
	}
	if rep.MCPConfigTokens > 0 {
		if _, err := fmt.Fprintf(w, "mcp config %d tokens (config only, not in total or budget)\n", rep.MCPConfigTokens); err != nil {
			return err
		}
	}
	limit := len(rep.Rows)
	if limit > 5 {
		limit = 5
	}
	for i := 0; i < limit; i++ {
		row := rep.Rows[i]
		note := ""
		if row.Note != "" {
			note = " (" + row.Note + ")"
		}
		if _, err := fmt.Fprintf(w, "  %d %s %q%s\n", row.Tokens, row.Kind, row.Name, note); err != nil {
			return err
		}
	}
	for _, suggestion := range rep.Suggestions {
		if _, err := fmt.Fprintf(w, "suggest: %s\n", safe(suggestion)); err != nil {
			return err
		}
	}
	return nil
}

// JSON writes the report. It does not include schema bodies.
func JSON(w io.Writer, rep estimate.Report) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(rep)
}

// safe escapes control characters so a name cannot forge lines or terminal escapes.
func safe(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r < 0x20 || r == 0x7f {
			fmt.Fprintf(&b, "\\x%02x", r)
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}
