package config

import (
	"encoding/json"
	"testing"
)

func TestSavedQueryRoundTrip(t *testing.T) {
	cfg := Config{
		Vault: "/tmp/vault",
		Queries: []SavedQuery{
			{Name: "Drafts", Filter: "status:draft"},
			{Name: "Project Alpha", Filter: "#project/alpha"},
		},
	}

	raw, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var got Config
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if len(got.Queries) != 2 {
		t.Fatalf("expected 2 queries, got %d", len(got.Queries))
	}
	if got.Queries[0].Name != "Drafts" || got.Queries[0].Filter != "status:draft" {
		t.Fatalf("unexpected first query: %+v", got.Queries[0])
	}
	if got.Queries[1].Name != "Project Alpha" || got.Queries[1].Filter != "#project/alpha" {
		t.Fatalf("unexpected second query: %+v", got.Queries[1])
	}
}

func TestEmptyQueriesOmittedFromJSON(t *testing.T) {
	cfg := Config{Vault: "/tmp/vault"}

	raw, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatalf("unmarshal to map: %v", err)
	}

	if _, ok := m["queries"]; ok {
		t.Fatalf("expected queries key to be omitted when empty, but it was present")
	}
}
