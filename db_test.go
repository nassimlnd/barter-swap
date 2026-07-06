package main

import (
	"io"
	"log/slog"
	"testing"
)

func TestRunMigrationsMissingDir(t *testing.T) {
	t.Parallel()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	if err := runMigrations(t.Context(), nil, "does-not-exist-for-barterswap-tests", logger); err != nil {
		t.Fatalf("missing dir should be ignored, got %v", err)
	}
}
