package cmd

import (
	"testing"
)

func TestZonesRenameCommand_Exists(t *testing.T) {
	cmd, _, err := zonesCmd.Find([]string{"rename"})
	if err != nil {
		t.Fatalf("rename command not found: %v", err)
	}
	if cmd.Name() != "rename" {
		t.Errorf("expected command name 'rename', got '%s'", cmd.Name())
	}
}

func TestZonesRenameCommand_RequiresTwoArgs(t *testing.T) {
	cmd, _, _ := zonesCmd.Find([]string{"rename"})

	if cmd.Args == nil {
		t.Fatal("expected Args validator to be set")
	}

	// Test with wrong number of args
	err := cmd.Args(cmd, []string{})
	if err == nil {
		t.Error("expected error with 0 args")
	}

	err = cmd.Args(cmd, []string{"one"})
	if err == nil {
		t.Error("expected error with 1 arg")
	}

	err = cmd.Args(cmd, []string{"one", "two", "three"})
	if err == nil {
		t.Error("expected error with 3 args")
	}

	// Test with correct number of args
	err = cmd.Args(cmd, []string{"zone-name", "new-name"})
	if err != nil {
		t.Errorf("expected no error with 2 args, got: %v", err)
	}
}

func TestZonesRenameCommand_Usage(t *testing.T) {
	cmd, _, _ := zonesCmd.Find([]string{"rename"})

	expected := "rename <name-or-id> <new-name>"
	if cmd.Use != expected {
		t.Errorf("expected Use '%s', got '%s'", expected, cmd.Use)
	}
}

func TestZonesDeleteCommand_Exists(t *testing.T) {
	cmd, _, err := zonesCmd.Find([]string{"delete"})
	if err != nil {
		t.Fatalf("delete command not found: %v", err)
	}
	if cmd.Name() != "delete" {
		t.Errorf("expected command name 'delete', got '%s'", cmd.Name())
	}
}

func TestZonesListCommand_Exists(t *testing.T) {
	cmd, _, err := zonesCmd.Find([]string{"list"})
	if err != nil {
		t.Fatalf("list command not found: %v", err)
	}
	if cmd.Name() != "list" {
		t.Errorf("expected command name 'list', got '%s'", cmd.Name())
	}
}
