package main

import (
	"log/slog"
	"os"

	"github.com/taylormonacelli/bluecare"
	"github.com/taylormonacelli/goldbug"
)

func main() {
	goldbug.SetDefaultLoggerText(slog.LevelDebug)

	status := bluecare.Execute("ec2", "us-west-2")
	os.Exit(status)
}
