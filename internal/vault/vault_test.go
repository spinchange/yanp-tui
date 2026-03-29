package vault

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestResolvePrecedence(t *testing.T) {
	v, err := Load(filepath.Join("..", "..", "testdata", "sample-vault"))
	if err != nil {
		t.Fatalf("load vault: %v", err)
	}

	result := v.Resolve("Sprint Review")
	if !result.Resolved || result.Note.RelPath != "projects/sprint-review.md" {
		t.Fatalf("expected title match for sprint review, got %+v", result)
	}

	result = v.Resolve("Friday Review")
	if !result.Resolved || result.Matched != "alias" {
		t.Fatalf("expected alias match, got %+v", result)
	}

	result = v.Resolve("alice")
	if !result.Resolved || result.Matched != "title" {
		t.Fatalf("expected case-insensitive title match, got %+v", result)
	}
}

func TestExtractInlineTagsSkipsCode(t *testing.T) {
	body := strings.Join([]string{
		"Meeting note with #planning and #topic/subtopic.",
		"`inline #ignore` but keep #work.",
		"```",
		"#also-ignore",
		"```",
		"<div>",
		"#html-ignore",
		"</div>",
	}, "\n")

	tags := ExtractInlineTags(body)
	if len(tags) != 3 {
		t.Fatalf("expected 3 tags, got %v", tags)
	}
}

func TestPublishTransformsWikilinks(t *testing.T) {
	root := filepath.Join("..", "..", "testdata", "sample-vault")
	v, err := Load(root)
	if err != nil {
		t.Fatalf("load vault: %v", err)
	}

	out := t.TempDir()
	warnings, err := v.Publish(PublishOptions{
		OutputDir:           out,
		MarkUnresolved:      true,
		PreserveFrontmatter: true,
	})
	if err != nil {
		t.Fatalf("publish: %v", err)
	}
	if len(warnings) != 0 {
		t.Fatalf("expected no warnings, got %v", warnings)
	}

	raw, err := os.ReadFile(filepath.Join(out, "projects", "sprint-review.md"))
	if err != nil {
		t.Fatalf("read published note: %v", err)
	}

	content := string(raw)
	if strings.Contains(content, "[[") {
		t.Fatalf("expected wikilinks to be rewritten, got %s", content)
	}
	if !strings.Contains(content, "[Alice](../alice.md)") {
		t.Fatalf("expected converted relative link, got %s", content)
	}
}

func TestRenameRewritesInboundLinks(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "alpha.md", "---\ntitle: Alpha\n---\n\nSee [[Beta]].\n")
	writeFixture(t, root, "beta.md", "---\ntitle: Beta\naliases:\n  - Old Beta\n---\n\n# Beta\n")

	v, err := Load(root)
	if err != nil {
		t.Fatalf("load vault: %v", err)
	}

	newPath, warnings, err := v.RenameNote("beta.md", "Gamma")
	if err != nil {
		t.Fatalf("rename note: %v", err)
	}
	if len(warnings) != 0 {
		t.Fatalf("expected no warnings, got %v", warnings)
	}
	if newPath != "gamma.md" {
		t.Fatalf("expected gamma.md, got %s", newPath)
	}

	raw, err := os.ReadFile(filepath.Join(root, "alpha.md"))
	if err != nil {
		t.Fatalf("read alpha: %v", err)
	}
	if !strings.Contains(string(raw), "[[Gamma]]") {
		t.Fatalf("expected inbound links rewritten, got %s", string(raw))
	}
}

func TestResolveConflictByTitleUsesMostRecentNote(t *testing.T) {
	root := t.TempDir()
	writeFixtureAt(t, root, "older.md", "---\ntitle: Shared\n---\n\nold\n", time.Date(2026, 3, 28, 10, 0, 0, 0, time.UTC))
	writeFixtureAt(t, root, "newer.md", "---\ntitle: Shared\n---\n\nnew\n", time.Date(2026, 3, 28, 11, 0, 0, 0, time.UTC))

	v, err := Load(root)
	if err != nil {
		t.Fatalf("load vault: %v", err)
	}

	result := v.Resolve("Shared")
	if !result.Resolved {
		t.Fatalf("expected conflict to resolve deterministically")
	}
	if result.Note.RelPath != "newer.md" {
		t.Fatalf("expected newer.md to win, got %s", result.Note.RelPath)
	}
	if !strings.Contains(result.Warning, "title conflict") {
		t.Fatalf("expected title conflict warning, got %q", result.Warning)
	}
}

func TestResolveConflictByAliasUsesMostRecentNote(t *testing.T) {
	root := t.TempDir()
	writeFixtureAt(t, root, "alpha.md", "---\ntitle: Alpha\naliases:\n  - Shared Alias\n---\n", time.Date(2026, 3, 28, 9, 0, 0, 0, time.UTC))
	writeFixtureAt(t, root, "beta.md", "---\ntitle: Beta\naliases:\n  - Shared Alias\n---\n", time.Date(2026, 3, 28, 12, 0, 0, 0, time.UTC))

	v, err := Load(root)
	if err != nil {
		t.Fatalf("load vault: %v", err)
	}

	result := v.Resolve("Shared Alias")
	if !result.Resolved || result.Note.RelPath != "beta.md" {
		t.Fatalf("expected beta.md to win alias conflict, got %+v", result)
	}
	if result.Matched != "alias" {
		t.Fatalf("expected alias match, got %+v", result)
	}
	if !strings.Contains(result.Warning, "alias conflict") {
		t.Fatalf("expected alias conflict warning, got %q", result.Warning)
	}
}

func TestResolveConflictByFilenameUsesMostRecentNote(t *testing.T) {
	root := t.TempDir()
	writeFixtureAt(t, root, "folder-one/shared.md", "# one\n", time.Date(2026, 3, 28, 8, 0, 0, 0, time.UTC))
	writeFixtureAt(t, root, "folder-two/shared.md", "# two\n", time.Date(2026, 3, 28, 13, 0, 0, 0, time.UTC))

	v, err := Load(root)
	if err != nil {
		t.Fatalf("load vault: %v", err)
	}

	result := v.Resolve("shared")
	if !result.Resolved || result.Note.RelPath != "folder-two/shared.md" {
		t.Fatalf("expected newer filename match to win, got %+v", result)
	}
	if result.Matched != "filename" {
		t.Fatalf("expected filename match, got %+v", result)
	}
	if !strings.Contains(result.Warning, "filename conflict") {
		t.Fatalf("expected filename conflict warning, got %q", result.Warning)
	}
}

func TestConflictsReportsAllDuplicateTargets(t *testing.T) {
	root := t.TempDir()
	writeFixtureAt(t, root, "a.md", "---\ntitle: Shared Title\naliases:\n  - Shared Alias\n---\n", time.Date(2026, 3, 28, 9, 0, 0, 0, time.UTC))
	writeFixtureAt(t, root, "b.md", "---\ntitle: Shared Title\naliases:\n  - Shared Alias\n---\n", time.Date(2026, 3, 28, 10, 0, 0, 0, time.UTC))
	writeFixtureAt(t, root, "folder-one/shared.md", "# one\n", time.Date(2026, 3, 28, 11, 0, 0, 0, time.UTC))
	writeFixtureAt(t, root, "folder-two/shared.md", "# two\n", time.Date(2026, 3, 28, 12, 0, 0, 0, time.UTC))

	v, err := Load(root)
	if err != nil {
		t.Fatalf("load vault: %v", err)
	}

	conflicts := v.Conflicts()
	if len(conflicts) != 3 {
		t.Fatalf("expected 3 conflicts, got %d", len(conflicts))
	}
	if conflicts[0].Winner == nil || len(conflicts[0].Candidates) < 2 {
		t.Fatalf("expected conflict winners and candidates, got %+v", conflicts[0])
	}
}

func writeFixture(t *testing.T, root, relPath, content string) {
	t.Helper()
	fullPath := filepath.Join(root, filepath.FromSlash(relPath))
	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(fullPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
}

func writeFixtureAt(t *testing.T, root, relPath, content string, modifiedAt time.Time) {
	t.Helper()
	writeFixture(t, root, relPath, content)
	fullPath := filepath.Join(root, filepath.FromSlash(relPath))
	if err := os.Chtimes(fullPath, modifiedAt, modifiedAt); err != nil {
		t.Fatalf("chtimes: %v", err)
	}
}
