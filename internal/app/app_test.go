package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spinchange/yanp-tui/internal/config"
)

// makeVault creates a minimal temp vault with a couple of notes and returns
// its path and a config pointing at it.
func makeVault(t *testing.T) (string, config.Config) {
	t.Helper()
	root := t.TempDir()

	write := func(rel, content string) {
		t.Helper()
		full := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
			t.Fatalf("write %s: %v", rel, err)
		}
	}

	write("alpha.md", "---\ntitle: Alpha\n---\n\nSee [[Beta]].\n")
	write("beta.md", "---\ntitle: Beta\n---\n\n# Beta\n")
	write("inbox.md", "# Inbox\n\n")

	return root, config.Config{Vault: root}
}

// --- capture ---

func TestRunCaptureAppendsToInbox(t *testing.T) {
	root, cfg := makeVault(t)

	err := Run([]string{"capture", "-text", "hello world"}, cfg)
	if err != nil {
		t.Fatalf("capture: %v", err)
	}

	raw, err := os.ReadFile(filepath.Join(root, "inbox.md"))
	if err != nil {
		t.Fatalf("read inbox: %v", err)
	}
	if !strings.Contains(string(raw), "hello world") {
		t.Fatalf("expected capture text in inbox.md, got:\n%s", raw)
	}
}

func TestRunCaptureRequiresText(t *testing.T) {
	_, cfg := makeVault(t)

	err := Run([]string{"capture", "-text", "   "}, cfg)
	if err == nil {
		t.Fatalf("expected error for blank capture text")
	}
}

func TestRunCaptureCreatesInboxIfMissing(t *testing.T) {
	root, cfg := makeVault(t)
	os.Remove(filepath.Join(root, "inbox.md"))

	err := Run([]string{"capture", "-text", "new entry"}, cfg)
	if err != nil {
		t.Fatalf("capture to missing inbox: %v", err)
	}

	raw, err := os.ReadFile(filepath.Join(root, "inbox.md"))
	if err != nil {
		t.Fatalf("read inbox: %v", err)
	}
	if !strings.Contains(string(raw), "new entry") {
		t.Fatalf("expected capture text in new inbox.md, got:\n%s", raw)
	}
}

// --- rename ---

func TestRunRenameRenamesNote(t *testing.T) {
	root, cfg := makeVault(t)

	err := Run([]string{"rename", "-note", "beta.md", "-title", "Gamma"}, cfg)
	if err != nil {
		t.Fatalf("rename: %v", err)
	}

	if _, err := os.Stat(filepath.Join(root, "gamma.md")); err != nil {
		t.Fatalf("expected gamma.md to exist after rename: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "beta.md")); err == nil {
		t.Fatalf("expected beta.md to be gone after rename")
	}
}

func TestRunRenameRewritesInboundLinks(t *testing.T) {
	root, cfg := makeVault(t)

	err := Run([]string{"rename", "-note", "beta.md", "-title", "Gamma"}, cfg)
	if err != nil {
		t.Fatalf("rename: %v", err)
	}

	raw, err := os.ReadFile(filepath.Join(root, "alpha.md"))
	if err != nil {
		t.Fatalf("read alpha: %v", err)
	}
	if !strings.Contains(string(raw), "[[Gamma]]") {
		t.Fatalf("expected inbound link rewritten in alpha.md, got:\n%s", raw)
	}
}

func TestRunRenameRequiresBothFlags(t *testing.T) {
	_, cfg := makeVault(t)

	if err := Run([]string{"rename", "-note", "beta.md"}, cfg); err == nil {
		t.Fatalf("expected error when -title is missing")
	}
	if err := Run([]string{"rename", "-title", "Gamma"}, cfg); err == nil {
		t.Fatalf("expected error when -note is missing")
	}
}

func TestRunRenameRejectsConflict(t *testing.T) {
	_, cfg := makeVault(t)

	err := Run([]string{"rename", "-note", "alpha.md", "-title", "Beta"}, cfg)
	if err == nil {
		t.Fatalf("expected error when renaming to an existing note's slug")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("expected 'already exists' error, got: %v", err)
	}
}

// --- publish ---

func TestRunPublishCreatesOutputFiles(t *testing.T) {
	root, cfg := makeVault(t)
	out := filepath.Join(root, "_pub")

	err := Run([]string{"publish", "-out", out}, cfg)
	if err != nil {
		t.Fatalf("publish: %v", err)
	}

	for _, rel := range []string{"alpha.md", "beta.md", "inbox.md"} {
		if _, err := os.Stat(filepath.Join(out, rel)); err != nil {
			t.Fatalf("expected %s in publish output: %v", rel, err)
		}
	}
}

func TestRunPublishRewritesWikilinks(t *testing.T) {
	root, cfg := makeVault(t)
	out := filepath.Join(root, "_pub")

	err := Run([]string{"publish", "-out", out}, cfg)
	if err != nil {
		t.Fatalf("publish: %v", err)
	}

	raw, err := os.ReadFile(filepath.Join(out, "alpha.md"))
	if err != nil {
		t.Fatalf("read published alpha: %v", err)
	}
	if strings.Contains(string(raw), "[[") {
		t.Fatalf("expected wikilinks rewritten in published output, got:\n%s", raw)
	}
	if !strings.Contains(string(raw), "[Beta](beta.md)") {
		t.Fatalf("expected markdown link in published alpha.md, got:\n%s", raw)
	}
}

func TestRunPublishSkipsDrafts(t *testing.T) {
	root, cfg := makeVault(t)
	draftPath := filepath.Join(root, "draft-note.md")
	if err := os.WriteFile(draftPath, []byte("---\ntitle: Draft Note\nstatus: draft\n---\n"), 0o644); err != nil {
		t.Fatalf("write draft: %v", err)
	}
	out := filepath.Join(root, "_pub")

	err := Run([]string{"publish", "-out", out, "-skip-drafts"}, cfg)
	if err != nil {
		t.Fatalf("publish: %v", err)
	}

	if _, err := os.Stat(filepath.Join(out, "draft-note.md")); err == nil {
		t.Fatalf("expected draft-note.md to be skipped in published output")
	}
}

// --- unknown command ---

func TestRunUnknownCommandReturnsError(t *testing.T) {
	_, cfg := makeVault(t)
	err := Run([]string{"frobnicate"}, cfg)
	if err == nil {
		t.Fatalf("expected error for unknown command")
	}
	if !strings.Contains(err.Error(), "frobnicate") {
		t.Fatalf("expected command name in error, got: %v", err)
	}
}
