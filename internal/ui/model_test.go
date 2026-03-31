package ui

import (
	"testing"
	"time"

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

func TestUnresolvedLinkCount(t *testing.T) {
	links := []vault.UnresolvedLink{
		{Target: "Missing One", Count: 2},
		{Target: "Missing Two", Count: 1},
	}
	if got := unresolvedLinkCount(links); got != 3 {
		t.Fatalf("expected 3 unresolved links, got %d", got)
	}
}
