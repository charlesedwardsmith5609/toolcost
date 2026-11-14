package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	"toolcost/internal/estimate"
	"toolcost/internal/load"
	"toolcost/internal/report"
)

func main() {
	format := flag.String("format", "text", "text or json")
	budget := flag.String("budget", "", "exit non-zero when the per-turn estimate is greater than this many tokens")
	flag.Parse()
	if flag.NArg() == 0 {
		fmt.Fprintln(os.Stderr, "usage: toolcost [--format text|json] [--budget N] file...")
		os.Exit(2)
	}
	limit := -1
	if *budget != "" {
		parsed, err := strconv.Atoi(*budget)
		if err != nil || parsed < 0 {
			fmt.Fprintln(os.Stderr, "budget must be a non-negative integer")
			os.Exit(2)
		}
		limit = parsed
	}
	var items []load.Item
	for _, path := range flag.Args() {
		parsed, err := load.ReadFile(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "read %s: %v\n", path, err)
			os.Exit(2)
		}
		items = append(items, parsed...)
	}
	rep, err := estimate.Count(items, estimate.Default())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	if limit >= 0 {
		rep.Budget = limit
	}
	if *format == "json" {
		err = report.JSON(os.Stdout, rep)
	} else if *format == "text" {
		err = report.Text(os.Stdout, rep)
	} else {
		fmt.Fprintln(os.Stderr, "format must be text or json")
		os.Exit(2)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	if limit >= 0 && rep.Total > limit {
		os.Exit(1)
	}
}
