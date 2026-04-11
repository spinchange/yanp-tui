package ui

import (
	"testing"

	"github.com/spinchange/yanp-tui/internal/config"
)

func TestResolveEditor(t *testing.T) {
	tests := []struct {
		name     string
		editor   string // cfg.Editor value
		envValue string // EDITOR env var value ("" means unset)
		setEnv   bool   // whether to set the EDITOR env var at all
		want     string
	}{
		{
			name:   "cfg.Editor set returns cfg value",
			editor: "vim",
			setEnv: false,
			want:   "vim",
		},
		{
			name:     "cfg.Editor empty, EDITOR env set returns env value",
			editor:   "",
			envValue: "nano",
			setEnv:   true,
			want:     "nano",
		},
		{
			name:   "cfg.Editor empty, EDITOR env unset returns notepad.exe",
			editor: "",
			setEnv: false,
			want:   "notepad.exe",
		},
		{
			name:   "cfg.Editor with surrounding whitespace returns trimmed value",
			editor: "  code  ",
			setEnv: false,
			want:   "code",
		},
		{
			name:     "cfg.Editor whitespace-only falls through to EDITOR env",
			editor:   "   ",
			envValue: "emacs",
			setEnv:   true,
			want:     "emacs",
		},
		{
			name:   "cfg.Editor whitespace-only and no env returns notepad.exe",
			editor: "   ",
			setEnv: false,
			want:   "notepad.exe",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Always clear EDITOR first so tests are hermetic.
			t.Setenv("EDITOR", "")
			if tc.setEnv {
				t.Setenv("EDITOR", tc.envValue)
			}

			cfg := config.Config{Editor: tc.editor}
			got := resolveEditor(cfg)
			if got != tc.want {
				t.Errorf("resolveEditor() = %q, want %q", got, tc.want)
			}
		})
	}
}
