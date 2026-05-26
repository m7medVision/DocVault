// Package main is the entry point for the DocVault Core Platform Service.
package main

import (
	"log/slog"
	"os"

	"github.com/docvault/backend/internal/app"
)

func main() {
	if err := app.Run(); err != nil {
		slog.Error("backend stopped", "error", err)
		os.Exit(1)
	}
}
