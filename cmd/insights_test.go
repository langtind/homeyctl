package cmd

import (
	"testing"
)

func TestResolveInsightLog_UsesFullID(t *testing.T) {
	logs := []InsightLog{
		{
			ID:       "homey:device:abc-123:measure_temperature",
			OwnerURI: "homey:device:abc-123",
			OwnerID:  "measure_temperature",
			Title:    "System temperature",
		},
	}

	// The Homey API looks up logs by their full ID, not by ownerId.
	uri, id, title, found := resolveInsightLog(logs, "homey:device:abc-123:measure_temperature")
	if !found {
		t.Fatal("expected log to be found")
	}
	if id != "homey:device:abc-123:measure_temperature" {
		t.Errorf("expected id to be the full log ID, got %q", id)
	}
	if uri != "homey:device:abc-123" {
		t.Errorf("expected uri %q, got %q", "homey:device:abc-123", uri)
	}
	if title != "System temperature" {
		t.Errorf("unexpected title %q", title)
	}
}

func TestResolveInsightLog_NotFound(t *testing.T) {
	_, _, _, found := resolveInsightLog(nil, "missing")
	if found {
		t.Error("expected not found for empty log list")
	}
}

func TestParseInsightEntries_ObjectWithValues(t *testing.T) {
	// The Homey insights API returns an object wrapping a "values" array,
	// not a flat array of entries.
	raw := []byte(`{
		"start": "2026-06-20T21:00:00.000Z",
		"end": "2026-06-20T22:00:00.000Z",
		"values": [
			{"t": "2026-06-20T21:35:00.000Z", "v": 24.7},
			{"t": "2026-06-20T21:40:00.000Z", "v": 24.8}
		]
	}`)

	entries, err := parseInsightEntries(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if entries[0].V != 24.7 {
		t.Errorf("expected first value 24.7, got %v", entries[0].V)
	}
}

func TestInsightsListCommand_Exists(t *testing.T) {
	cmd, _, err := insightsCmd.Find([]string{"list"})
	if err != nil {
		t.Fatalf("list command not found: %v", err)
	}
	if cmd.Name() != "list" {
		t.Errorf("expected command name 'list', got '%s'", cmd.Name())
	}
}

func TestInsightsGetCommand_Exists(t *testing.T) {
	cmd, _, err := insightsCmd.Find([]string{"get"})
	if err != nil {
		t.Fatalf("get command not found: %v", err)
	}
	if cmd.Name() != "get" {
		t.Errorf("expected command name 'get', got '%s'", cmd.Name())
	}
}

func TestInsightsGetCommand_RequiresOneArg(t *testing.T) {
	cmd, _, _ := insightsCmd.Find([]string{"get"})

	if cmd.Args == nil {
		t.Fatal("expected Args validator to be set")
	}

	err := cmd.Args(cmd, []string{})
	if err == nil {
		t.Error("expected error with 0 args")
	}

	err = cmd.Args(cmd, []string{"log-id"})
	if err != nil {
		t.Errorf("expected no error with 1 arg, got: %v", err)
	}
}

func TestInsightsDeleteCommand_Exists(t *testing.T) {
	cmd, _, err := insightsCmd.Find([]string{"delete"})
	if err != nil {
		t.Fatalf("delete command not found: %v", err)
	}
	if cmd.Name() != "delete" {
		t.Errorf("expected command name 'delete', got '%s'", cmd.Name())
	}
}

func TestInsightsDeleteCommand_RequiresOneArg(t *testing.T) {
	cmd, _, _ := insightsCmd.Find([]string{"delete"})

	if cmd.Args == nil {
		t.Fatal("expected Args validator to be set")
	}

	err := cmd.Args(cmd, []string{})
	if err == nil {
		t.Error("expected error with 0 args")
	}

	err = cmd.Args(cmd, []string{"log-id"})
	if err != nil {
		t.Errorf("expected no error with 1 arg, got: %v", err)
	}
}

func TestInsightsClearCommand_Exists(t *testing.T) {
	cmd, _, err := insightsCmd.Find([]string{"clear"})
	if err != nil {
		t.Fatalf("clear command not found: %v", err)
	}
	if cmd.Name() != "clear" {
		t.Errorf("expected command name 'clear', got '%s'", cmd.Name())
	}
}

func TestInsightsClearCommand_RequiresOneArg(t *testing.T) {
	cmd, _, _ := insightsCmd.Find([]string{"clear"})

	if cmd.Args == nil {
		t.Fatal("expected Args validator to be set")
	}

	err := cmd.Args(cmd, []string{})
	if err == nil {
		t.Error("expected error with 0 args")
	}

	err = cmd.Args(cmd, []string{"log-id"})
	if err != nil {
		t.Errorf("expected no error with 1 arg, got: %v", err)
	}
}
