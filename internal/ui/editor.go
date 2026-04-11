package ui

import (
	"os"
	"strings"

	"github.com/spinchange/yanp-tui/internal/config"
)

// resolveEditor returns the editor to use when opening notes.
// Priority order:
//  1. cfg.Editor (trimmed), if non-empty
//  2. EDITOR environment variable, if set and non-empty
//  3. "notepad.exe" as the default fallback
func resolveEditor(cfg config.Config) string {
	if trimmed := strings.TrimSpace(cfg.Editor); trimmed != "" {
		return trimmed
	}
	if env := os.Getenv("EDITOR"); env != "" {
		return env
	}
	return "notepad.exe"
}
