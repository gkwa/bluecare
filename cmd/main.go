package main

import (
	"log/slog"
	"os"

	"github.com/taylormonacelli/bluecare"
	"github.com/taylormonacelli/goldbug"
)

func main() {
	goldbug.SetDefaultLoggerText(slog.LevelDebug)

	status := bluecare.Execute()
	os.Exit(status)
}
