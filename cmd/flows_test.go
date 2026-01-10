package cmd

import (
	"strings"
	"testing"
)

func TestValidateFlow_MissingName(t *testing.T) {
	flow := map[string]interface{}{}

	err := validateFlow(flow, false)
	if err == nil {
		t.Error("expected error for missing name")
	}
	if !strings.Contains(err.Error(), "'name' is required") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidateFlow_EmptyName(t *testing.T) {
	flow := map[string]interface{}{
		"name": "",
	}

	err := validateFlow(flow, false)
	if err == nil {
		t.Error("expected error for empty name")
	}
}

func TestValidateFlow_MissingTrigger(t *testing.T) {
	flow := map[string]interface{}{
		"name": "Test Flow",
	}

	err := validateFlow(flow, false)
	if err == nil {
		t.Error("expected error for missing trigger")
	}
	if !strings.Contains(err.Error(), "'trigger' is required") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidateFlow_InvalidTriggerID(t *testing.T) {
	flow := map[string]interface{}{
		"name": "Test Flow",
		"trigger": map[string]interface{}{
			"id": "invalid-trigger",
		},
	}

	err := validateFlow(flow, false)
	if err == nil {
		t.Error("expected error for invalid trigger ID")
	}
	if !strings.Contains(err.Error(), "must start with 'homey:'") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidateFlow_WrongDroptokenFormat(t *testing.T) {
	flow := map[string]interface{}{
		"name": "Test Flow",
		"trigger": map[string]interface{}{
			"id": "homey:manager:presence:user_enter",
		},
		"conditions": []interface{}{
			map[string]interface{}{
				"id":        "homey:manager:logic:lt",
				"droptoken": "homey:device:abc123:measure_temperature", // Wrong: uses : instead of |
			},
		},
	}

	err := validateFlow(flow, false)
	if err == nil {
		t.Error("expected error for wrong droptoken format")
	}
	if !strings.Contains(err.Error(), "droptoken uses wrong format") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidateFlow_CorrectDroptokenFormat(t *testing.T) {
	flow := map[string]interface{}{
		"name": "Test Flow",
		"trigger": map[string]interface{}{
			"id": "homey:manager:presence:user_enter",
		},
		"conditions": []interface{}{
			map[string]interface{}{
				"id":        "homey:manager:logic:lt",
				"droptoken": "homey:device:abc123|measure_temperature", // Correct: uses |
			},
		},
	}

	err := validateFlow(flow, false)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidateFlow_ValidSimpleFlow(t *testing.T) {
	flow := map[string]interface{}{
		"name": "Test Flow",
		"trigger": map[string]interface{}{
			"id":   "homey:manager:presence:user_enter",
			"args": map[string]interface{}{},
		},
		"actions": []interface{}{
			map[string]interface{}{
				"id":   "homey:device:abc123:on",
				"args": map[string]interface{}{},
			},
		},
	}

	err := validateFlow(flow, false)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestNormalizeSimpleFlow_AddsGroupToActions(t *testing.T) {
	flow := map[string]interface{}{
		"actions": []interface{}{
			map[string]interface{}{
				"id": "homey:device:abc:on",
			},
		},
	}

	normalizeSimpleFlow(flow)

	actions := flow["actions"].([]interface{})
	action := actions[0].(map[string]interface{})

	if action["group"] != "then" {
		t.Errorf("expected group 'then', got %v", action["group"])
	}
}

func TestNormalizeSimpleFlow_AddsGroupToConditions(t *testing.T) {
	flow := map[string]interface{}{
		"conditions": []interface{}{
			map[string]interface{}{
				"id": "homey:device:abc:on",
			},
		},
	}

	normalizeSimpleFlow(flow)

	conditions := flow["conditions"].([]interface{})
	condition := conditions[0].(map[string]interface{})

	if condition["group"] != "group1" {
		t.Errorf("expected group 'group1', got %v", condition["group"])
	}
	if condition["inverted"] != false {
		t.Errorf("expected inverted false, got %v", condition["inverted"])
	}
}

func TestNormalizeSimpleFlow_PreservesExistingGroup(t *testing.T) {
	flow := map[string]interface{}{
		"actions": []interface{}{
			map[string]interface{}{
				"id":    "homey:device:abc:on",
				"group": "else",
			},
		},
	}

	normalizeSimpleFlow(flow)

	actions := flow["actions"].([]interface{})
	action := actions[0].(map[string]interface{})

	if action["group"] != "else" {
		t.Errorf("expected group 'else' to be preserved, got %v", action["group"])
	}
}
