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

func TestSlugifyReplacesPathSeparators(t *testing.T) {
	slug := Slugify("Foo / Bar\\Baz")
	if slug != "foo-bar-baz" {
		t.Fatalf("expected foo-bar-baz, got %q", slug)
	}
}

func TestSplitFrontmatterNormalizesCRLF(t *testing.T) {
	raw := []byte("---\r\ntitle: Test\r\naliases:\r\n  - Alias\r\n---\r\n\r\nBody\r\n")

	fm, body, ok := splitFrontmatter(raw)
	if !ok {
		t.Fatalf("expected frontmatter to be parsed")
	}
	if strings.Contains(string(fm), "\r") {
		t.Fatalf("expected normalized frontmatter, got %q", string(fm))
	}
	if strings.Contains(string(body), "\r") {
		t.Fatalf("expected normalized body, got %q", string(body))
	}
	if string(body) != "\nBody\n" {
		t.Fatalf("expected normalized body content, got %q", string(body))
	}
}

func TestExtractInlineTagsHandlesEscapedAndDoubleBackticks(t *testing.T) {
	body := "Escaped \\`marker and #keep plus ``#ignore`` and `#also-ignore` and #work."

	tags := ExtractInlineTags(body)
	if len(tags) != 2 {
		t.Fatalf("expected 2 tags, got %v", tags)
	}
	if tags[0] != "keep" || tags[1] != "work" {
		t.Fatalf("expected [keep work], got %v", tags)
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

func TestUnresolvedLinksReportsUniqueNoteTargets(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "alpha.md", strings.Join([]string{
		"# Alpha",
		"",
		"See [[Missing Note]] and [[Missing Note|alias]].",
		"`[[Inline Code Missing]]`",
		"```",
		"[[Fenced Missing]]",
		"```",
		"Resolved link to [[Beta]].",
		"",
	}, "\n"))
	writeFixture(t, root, "beta.md", "---\ntitle: Beta\n---\n")
	writeFixture(t, root, "gamma.md", "Also see [[Another Missing]].\n")

	v, err := Load(root)
	if err != nil {
		t.Fatalf("load vault: %v", err)
	}

	unresolved := v.UnresolvedLinks()
	if len(unresolved) != 2 {
		t.Fatalf("expected 2 unresolved note-target pairs, got %d (%+v)", len(unresolved), unresolved)
	}

	if unresolved[0].Source.RelPath != "alpha.md" || unresolved[0].Target != "Missing Note" || unresolved[0].Count != 2 {
		t.Fatalf("unexpected first unresolved link: %+v", unresolved[0])
	}
	if unresolved[1].Source.RelPath != "gamma.md" || unresolved[1].Target != "Another Missing" || unresolved[1].Count != 1 {
		t.Fatalf("unexpected second unresolved link: %+v", unresolved[1])
	}
}

func TestEnsurePeriodicDailyNoteCreatesAndReuses(t *testing.T) {
	root := t.TempDir()
	v, err := Load(root)
	if err != nil {
		t.Fatalf("load vault: %v", err)
	}

	when := time.Date(2026, 3, 30, 9, 0, 0, 0, time.Local)
	note, created, err := v.EnsurePeriodicNote(PeriodicDaily, when)
	if err != nil {
		t.Fatalf("ensure daily note: %v", err)
	}
	if !created {
		t.Fatalf("expected daily note to be created")
	}
	if note.RelPath != "daily/2026-03-30.md" {
		t.Fatalf("unexpected daily path: %s", note.RelPath)
	}

	noteAgain, createdAgain, err := v.EnsurePeriodicNote(PeriodicDaily, when)
	if err != nil {
		t.Fatalf("reuse daily note: %v", err)
	}
	if createdAgain {
		t.Fatalf("expected existing daily note to be reused")
	}
	if noteAgain.RelPath != note.RelPath {
		t.Fatalf("expected same note path, got %s", noteAgain.RelPath)
	}
}

func TestEnsurePeriodicWeeklyNoteUsesIsoWeekPath(t *testing.T) {
	root := t.TempDir()
	v, err := Load(root)
	if err != nil {
		t.Fatalf("load vault: %v", err)
	}

	when := time.Date(2026, 3, 30, 9, 0, 0, 0, time.Local)
	note, created, err := v.EnsurePeriodicNote(PeriodicWeekly, when)
	if err != nil {
		t.Fatalf("ensure weekly note: %v", err)
	}
	if !created {
		t.Fatalf("expected weekly note to be created")
	}
	if note.RelPath != "weekly/2026-W14.md" {
		t.Fatalf("unexpected weekly path: %s", note.RelPath)
	}
}

func TestEnsurePeriodicMonthlyNoteUsesMonthPath(t *testing.T) {
	root := t.TempDir()
	v, err := Load(root)
	if err != nil {
		t.Fatalf("load vault: %v", err)
	}

	when := time.Date(2026, 3, 30, 9, 0, 0, 0, time.Local)
	note, created, err := v.EnsurePeriodicNote(PeriodicMonthly, when)
	if err != nil {
		t.Fatalf("ensure monthly note: %v", err)
	}
	if !created {
		t.Fatalf("expected monthly note to be created")
	}
	if note.RelPath != "monthly/2026-03.md" {
		t.Fatalf("unexpected monthly path: %s", note.RelPath)
	}
}

func TestRenameRejectsConflictingDestination(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "alpha.md", "---\ntitle: Alpha\n---\n")
	writeFixture(t, root, "gamma.md", "---\ntitle: Gamma\n---\n")

	v, err := Load(root)
	if err != nil {
		t.Fatalf("load vault: %v", err)
	}

	_, _, err = v.RenameNote("alpha.md", "Gamma")
	if err == nil {
		t.Fatalf("expected error when renaming to an existing note's slug")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("expected 'already exists' error, got: %v", err)
	}

	// alpha.md should be unchanged
	raw, readErr := os.ReadFile(filepath.Join(root, "alpha.md"))
	if readErr != nil {
		t.Fatalf("read alpha: %v", readErr)
	}
	if !strings.Contains(string(raw), "Alpha") {
		t.Fatalf("expected alpha.md to be unchanged after failed rename")
	}
}

func TestRenameRejectsSameTitle(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "alpha.md", "---\ntitle: Alpha\n---\n")

	v, err := Load(root)
	if err != nil {
		t.Fatalf("load vault: %v", err)
	}

	_, _, err = v.RenameNote("alpha.md", "Alpha")
	if err == nil {
		t.Fatalf("expected error when renaming to the same filename")
	}
	if !strings.Contains(err.Error(), "same filename") {
		t.Fatalf("expected 'same filename' error, got: %v", err)
	}
}

func TestMalformedFrontmatterIsRecordedNotFatal(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "good.md", "---\ntitle: Good\n---\n\nBody.\n")
	writeFixture(t, root, "bad.md", "---\ntitle: [unclosed\n---\n\nBody.\n")
	writeFixture(t, root, "no-fm.md", "# No frontmatter\n")

	v, err := Load(root)
	if err != nil {
		t.Fatalf("load should succeed even with malformed frontmatter: %v", err)
	}
	if len(v.Notes) != 3 {
		t.Fatalf("expected all 3 notes indexed, got %d", len(v.Notes))
	}
	if len(v.ParseErrors) != 1 {
		t.Fatalf("expected 1 parse error, got %d", len(v.ParseErrors))
	}
	if v.ParseErrors[0].RelPath != "bad.md" {
		t.Fatalf("expected parse error for bad.md, got %q", v.ParseErrors[0].RelPath)
	}
	if v.ParseErrors[0].Err == nil {
		t.Fatalf("expected non-nil error in ParseError")
	}

	// bad.md should still be browsable with stem as title
	bad := v.NoteByRelPath("bad.md")
	if bad == nil {
		t.Fatalf("expected bad.md to be indexed")
	}
	if bad.Title != "bad" {
		t.Fatalf("expected stem as title for malformed note, got %q", bad.Title)
	}
}

func TestDraftNotes(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "draft-one.md", "---\ntitle: Draft One\nstatus: draft\n---\n")
	writeFixture(t, root, "draft-two.md", "---\ntitle: Draft Two\nstatus: Draft\n---\n") // case variation
	writeFixture(t, root, "active.md", "---\ntitle: Active\nstatus: active\n---\n")
	writeFixture(t, root, "no-status.md", "# No Status\n")

	v, err := Load(root)
	if err != nil {
		t.Fatalf("load vault: %v", err)
	}

	drafts := v.DraftNotes()
	if len(drafts) != 2 {
		t.Fatalf("expected 2 draft notes, got %d", len(drafts))
	}
	for _, d := range drafts {
		if !strings.EqualFold(d.Status, "draft") {
			t.Fatalf("expected draft status, got %q in %s", d.Status, d.RelPath)
		}
	}
}

func TestPeriodicNoteHasDateSpecificTitle(t *testing.T) {
	root := t.TempDir()
	v, err := Load(root)
	if err != nil {
		t.Fatalf("load vault: %v", err)
	}

	when := time.Date(2026, 4, 11, 9, 0, 0, 0, time.Local)

	dailyNote, _, err := v.EnsurePeriodicNote(PeriodicDaily, when)
	if err != nil {
		t.Fatalf("ensure daily: %v", err)
	}
	if dailyNote.Title != "2026-04-11" {
		t.Fatalf("expected daily title '2026-04-11', got %q", dailyNote.Title)
	}

	weeklyNote, _, err := v.EnsurePeriodicNote(PeriodicWeekly, when)
	if err != nil {
		t.Fatalf("ensure weekly: %v", err)
	}
	if weeklyNote.Title != "2026-W15" {
		t.Fatalf("expected weekly title '2026-W15', got %q", weeklyNote.Title)
	}

	monthlyNote, _, err := v.EnsurePeriodicNote(PeriodicMonthly, when)
	if err != nil {
		t.Fatalf("ensure monthly: %v", err)
	}
	if monthlyNote.Title != "2026-04" {
		t.Fatalf("expected monthly title '2026-04', got %q", monthlyNote.Title)
	}
}

func TestPeriodicRelPath(t *testing.T) {
	when := time.Date(2026, 4, 11, 9, 0, 0, 0, time.Local)

	daily, err := PeriodicRelPath(PeriodicDaily, when)
	if err != nil {
		t.Fatalf("daily rel path: %v", err)
	}
	if daily != "daily/2026-04-11.md" {
		t.Fatalf("expected daily/2026-04-11.md, got %q", daily)
	}

	weekly, err := PeriodicRelPath(PeriodicWeekly, when)
	if err != nil {
		t.Fatalf("weekly rel path: %v", err)
	}
	if weekly != "weekly/2026-W15.md" {
		t.Fatalf("expected weekly/2026-W15.md, got %q", weekly)
	}

	monthly, err := PeriodicRelPath(PeriodicMonthly, when)
	if err != nil {
		t.Fatalf("monthly rel path: %v", err)
	}
	if monthly != "monthly/2026-04.md" {
		t.Fatalf("expected monthly/2026-04.md, got %q", monthly)
	}
}

func TestStaleNotes(t *testing.T) {
	root := t.TempDir()
	asOf := time.Date(2026, 4, 11, 12, 0, 0, 0, time.UTC)
	recent := asOf.AddDate(0, 0, -10) // 10 days ago — not stale at 30 days
	stale := asOf.AddDate(0, 0, -45)  // 45 days ago — stale at 30 days

	writeFixtureAt(t, root, "new-note.md", "# New\n", recent)
	writeFixtureAt(t, root, "old-note.md", "# Old\n", stale)

	v, err := Load(root)
	if err != nil {
		t.Fatalf("load vault: %v", err)
	}

	got := v.StaleNotes(30, asOf)
	if len(got) != 1 {
		t.Fatalf("expected 1 stale note, got %d", len(got))
	}
	if got[0].RelPath != "old-note.md" {
		t.Fatalf("expected old-note.md to be stale, got %s", got[0].RelPath)
	}

	gotNone := v.StaleNotes(0, asOf)
	if len(gotNone) != 0 {
		t.Fatalf("expected no stale notes when days=0, got %d", len(gotNone))
	}
}

func TestInboundLinks(t *testing.T) {
	root := t.TempDir()
	// target note
	writeFixture(t, root, "target.md", "---\ntitle: Target Note\n---\n\n# Target\n")
	// two sources that link to target
	writeFixture(t, root, "source-a.md", "See [[Target Note]] for details.\n")
	writeFixture(t, root, "source-b.md", "Also check [[Target Note|the target]].\n")
	// a note with no link to target
	writeFixture(t, root, "unrelated.md", "Nothing here links to target.\n")
	// a note whose only link to target is inside a code fence — must not count
	writeFixture(t, root, "fenced.md", strings.Join([]string{
		"Some prose.",
		"```",
		"[[Target Note]]",
		"```",
		"End.",
		"",
	}, "\n"))

	v, err := Load(root)
	if err != nil {
		t.Fatalf("load vault: %v", err)
	}

	// Two inbound links from different sources
	got := v.InboundLinks("target.md")
	if len(got) != 2 {
		t.Fatalf("expected 2 inbound links, got %d: %v", len(got), got)
	}
	if got[0] != "source-a.md" || got[1] != "source-b.md" {
		t.Fatalf("expected [source-a.md source-b.md], got %v", got)
	}

	// No inbound links
	gotNone := v.InboundLinks("unrelated.md")
	if len(gotNone) != 0 {
		t.Fatalf("expected no inbound links for unrelated.md, got %v", gotNone)
	}

	// Wikilink inside code fence is not counted
	gotFenced := v.InboundLinks("fenced.md")
	if len(gotFenced) != 0 {
		t.Fatalf("expected no inbound links for fenced.md, got %v", gotFenced)
	}

	// Target note is not included in its own inbound links
	for _, p := range got {
		if p == "target.md" {
			t.Fatalf("target note must not appear in its own inbound links")
		}
	}
}

func TestDeleteNote(t *testing.T) {
	t.Run("happy path: note removed from disk and v.Notes", func(t *testing.T) {
		root := t.TempDir()
		writeFixture(t, root, "alpha.md", "---\ntitle: Alpha\n---\n\n# Alpha\n")
		writeFixture(t, root, "beta.md", "---\ntitle: Beta\n---\n\n# Beta\n")

		v, err := Load(root)
		if err != nil {
			t.Fatalf("load vault: %v", err)
		}
		if len(v.Notes) != 2 {
			t.Fatalf("expected 2 notes, got %d", len(v.Notes))
		}

		warnings, err := v.DeleteNote("alpha.md")
		if err != nil {
			t.Fatalf("delete note: %v", err)
		}
		if len(warnings) != 0 {
			t.Fatalf("expected no warnings, got %v", warnings)
		}

		// File must be gone from disk.
		if _, statErr := os.Stat(filepath.Join(root, "alpha.md")); statErr == nil {
			t.Fatalf("expected alpha.md to be deleted from disk")
		}

		// Note must be removed from v.Notes.
		if len(v.Notes) != 1 {
			t.Fatalf("expected 1 note after delete, got %d", len(v.Notes))
		}
		if v.NoteByRelPath("alpha.md") != nil {
			t.Fatalf("expected alpha.md to be removed from v.Notes")
		}
	})

	t.Run("inbound links: warnings returned, delete still succeeds", func(t *testing.T) {
		root := t.TempDir()
		writeFixture(t, root, "target.md", "---\ntitle: Target\n---\n\n# Target\n")
		writeFixture(t, root, "source-a.md", "See [[Target]] for details.\n")
		writeFixture(t, root, "source-b.md", "Also check [[Target|the target]].\n")

		v, err := Load(root)
		if err != nil {
			t.Fatalf("load vault: %v", err)
		}

		warnings, err := v.DeleteNote("target.md")
		if err != nil {
			t.Fatalf("delete note: %v", err)
		}
		if len(warnings) != 2 {
			t.Fatalf("expected 2 warnings, got %d: %v", len(warnings), warnings)
		}
		for _, w := range warnings {
			if !strings.Contains(w, "links to deleted note") {
				t.Fatalf("unexpected warning format: %q", w)
			}
		}

		// Delete still succeeded.
		if _, statErr := os.Stat(filepath.Join(root, "target.md")); statErr == nil {
			t.Fatalf("expected target.md to be deleted from disk")
		}
		if v.NoteByRelPath("target.md") != nil {
			t.Fatalf("expected target.md to be removed from v.Notes")
		}
	})

	t.Run("note not found: returns error containing 'not found'", func(t *testing.T) {
		root := t.TempDir()
		writeFixture(t, root, "alpha.md", "# Alpha\n")

		v, err := Load(root)
		if err != nil {
			t.Fatalf("load vault: %v", err)
		}

		_, err = v.DeleteNote("nonexistent.md")
		if err == nil {
			t.Fatalf("expected error for nonexistent note")
		}
		if !strings.Contains(err.Error(), "not found") {
			t.Fatalf("expected 'not found' in error, got: %v", err)
		}
	})
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
