package cmd

import (
	"testing"
)

func TestDevicesRenameCommand_Exists(t *testing.T) {
	// Verify the rename command is properly registered
	cmd, _, err := devicesCmd.Find([]string{"rename"})
	if err != nil {
		t.Fatalf("rename command not found: %v", err)
	}
	if cmd.Name() != "rename" {
		t.Errorf("expected command name 'rename', got '%s'", cmd.Name())
	}
}

func TestDevicesRenameCommand_RequiresTwoArgs(t *testing.T) {
	cmd, _, _ := devicesCmd.Find([]string{"rename"})

	// Test that the command requires exactly 2 args
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
	err = cmd.Args(cmd, []string{"device-name", "new-name"})
	if err != nil {
		t.Errorf("expected no error with 2 args, got: %v", err)
	}
}

func TestDevicesRenameCommand_Usage(t *testing.T) {
	cmd, _, _ := devicesCmd.Find([]string{"rename"})

	expected := "rename <name-or-id> <new-name>"
	if cmd.Use != expected {
		t.Errorf("expected Use '%s', got '%s'", expected, cmd.Use)
	}
}

func TestDevicesDeleteCommand_Exists(t *testing.T) {
	cmd, _, err := devicesCmd.Find([]string{"delete"})
	if err != nil {
		t.Fatalf("delete command not found: %v", err)
	}
	if cmd.Name() != "delete" {
		t.Errorf("expected command name 'delete', got '%s'", cmd.Name())
	}
}

func TestDevicesListCommand_Exists(t *testing.T) {
	cmd, _, err := devicesCmd.Find([]string{"list"})
	if err != nil {
		t.Fatalf("list command not found: %v", err)
	}
	if cmd.Name() != "list" {
		t.Errorf("expected command name 'list', got '%s'", cmd.Name())
	}
}
