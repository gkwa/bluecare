package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"sort"

	"github.com/taylormonacelli/bluecare"
	"github.com/taylormonacelli/goldbug"
)

var (
	verbose   bool
	services  bool
	logFormat string
)

func main() {
	flag.BoolVar(&verbose, "verbose", false, "Enable verbose output")
	flag.BoolVar(&verbose, "v", false, "Enable verbose output (shorthand)")
	flag.StringVar(&logFormat, "log-format", "", "Log format (text or json)")

	flag.BoolVar(&services, "services", false, "List services")
	flag.BoolVar(&services, "s", false, "List services (shorthand)")

	flag.Parse()

	if verbose || logFormat != "" {
		if logFormat == "json" {
			goldbug.SetDefaultLoggerJson(slog.LevelDebug)
		} else {
			goldbug.SetDefaultLoggerText(slog.LevelDebug)
		}
	}

	if services {
		services := bluecare.GetServices()
		sort.Strings(services)
		for _, service := range services {
			fmt.Fprintf(os.Stdout, "%s\n", service)
		}
		return
	}

	status := bluecare.Execute("ec2", "us-west-2")
	os.Exit(status)
}
