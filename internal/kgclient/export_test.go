package kgclient

import (
	"log/slog"
	"os"
	"testing"

	"github.com/cortexa-llc/ai-pack/internal/monitoring"
)

func TestMain(m *testing.M) {
	// Initialize the global logger so tests that exercise WriteBack
	// (which calls monitoring.Logger.*) don't panic with a nil pointer.
	monitoring.Logger = slog.New(slog.NewTextHandler(os.Stderr, nil))
	os.Exit(m.Run())
}
