package ui

import (
	"strings"
	"testing"
	"time"

	"github.com/spinchange/yanp-tui/internal/config"
	"github.com/spinchange/yanp-tui/internal/vault"
)

func TestInboxEntryCount(t *testing.T) {
	note := &vault.Note{
		Body: "# Inbox\n\n- one\n* two\nnot an item\n",
	}
	if got := inboxEntryCount(note); got != 2 {
		t.Fatalf("expected 2 inbox entries, got %d", got)
	}
}

func TestPeriodicSummary(t *testing.T) {
	when := time.Date(2026, 3, 30, 9, 0, 0, 0, time.Local)
	v := &vault.Vault{
		Notes: []*vault.Note{
			{RelPath: "daily/2026-03-30.md"},
		},
	}
	got := periodicSummary(v, vault.PeriodicDaily, when)
	if got != "daily/2026-03-30.md" {
		t.Fatalf("unexpected periodic summary: %s", got)
	}
}

func TestParseNoteInput(t *testing.T) {
	cases := []struct {
		input    string
		wantDir  string
		wantTitle string
	}{
		{"My Note", "", "My Note"},
		{"projects/My Note", "projects", "My Note"},
		{"a/b/My Note", "a/b", "My Note"},
		{"  projects/  My Note  ", "projects", "My Note"},
		{"/Leading Slash", "", "Leading Slash"},
		{"trailing/", "trailing", ""},
	}
	for _, c := range cases {
		dir, title := parseNoteInput(c.input)
		if dir != c.wantDir || title != c.wantTitle {
			t.Errorf("parseNoteInput(%q) = (%q, %q), want (%q, %q)",
				c.input, dir, title, c.wantDir, c.wantTitle)
		}
	}
}

func TestDashboardItemsIncludesSavedQueries(t *testing.T) {
	v := &vault.Vault{
		Notes: []*vault.Note{
			{RelPath: "alpha.md", Title: "Alpha", Tags: []string{"project"}},
			{RelPath: "beta.md", Title: "Beta"},
		},
	}
	cfg := config.Config{
		Queries: []config.SavedQuery{
			{Name: "Projects", Filter: "#project"},
			{Name: "Empty name", Filter: ""},
			{Name: "", Filter: "some filter"},
		},
	}
	m := Model{cfg: cfg, vault: v, allNotes: v.Notes, notes: v.Notes}

	items := m.dashboardItems()

	var queryItems []dashboardItem
	for _, item := range items {
		if item.action == "query" {
			queryItems = append(queryItems, item)
		}
	}
	if len(queryItems) != 1 {
		t.Fatalf("expected 1 query item (invalid entries skipped), got %d", len(queryItems))
	}
	if !strings.Contains(queryItems[0].label, "Projects") {
		t.Fatalf("expected label to contain query name, got %q", queryItems[0].label)
	}
	if queryItems[0].filter != "#project" {
		t.Fatalf("expected filter '#project', got %q", queryItems[0].filter)
	}
}

func TestUnresolvedLinkCount(t *testing.T) {
	links := []vault.UnresolvedLink{
		{Target: "Missing One", Count: 2},
		{Target: "Missing Two", Count: 1},
	}
	if got := unresolvedLinkCount(links); got != 3 {
		t.Fatalf("expected 3 unresolved links, got %d", got)
	}
}
