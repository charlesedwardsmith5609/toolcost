package report

import (
	"encoding/json"
	"fmt"
	"io"

	"toolcost/internal/estimate"
)

// Text writes the per-turn total, the top offenders, and suggestions.
func Text(w io.Writer, rep estimate.Report) error {
	if _, err := fmt.Fprintf(w, "total %d tokens per turn\n", rep.Total); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "estimator %s (%s)\n", rep.Estimator, rep.ErrorBound); err != nil {
		return err
	}
	limit := len(rep.Rows)
	if limit > 5 {
		limit = 5
	}
	for i := 0; i < limit; i++ {
		row := rep.Rows[i]
		if _, err := fmt.Fprintf(w, "  %d %s %s\n", row.Tokens, row.Kind, row.Name); err != nil {
			return err
		}
	}
	for _, suggestion := range rep.Suggestions {
		if _, err := fmt.Fprintf(w, "suggest: %s\n", suggestion); err != nil {
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
