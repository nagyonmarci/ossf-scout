package main

import (
	"testing"
)

func TestParseChecks(t *testing.T) {
	tests := []struct {
		input    string
		expected map[string]bool
	}{
		{"", map[string]bool{}},
		{"CI-Tests", map[string]bool{"CI-Tests": true}},
		{"CI-Tests,SAST", map[string]bool{"CI-Tests": true, "SAST": true}},
		{" CI-Tests , SAST ", map[string]bool{"CI-Tests": true, "SAST": true}},
		{",,,", map[string]bool{"": true}},
	}
	for _, tc := range tests {
		got := parseChecks(tc.input)
		for k := range tc.expected {
			if !got[k] {
				t.Errorf("parseChecks(%q): missing key %q", tc.input, k)
			}
		}
	}
}

func TestFindCheckScore(t *testing.T) {
	checks := []scorecardCheck{
		{Name: "CI-Tests", Score: 10},
		{Name: "SAST", Score: 7},
		{Name: "Branch-Protection", Score: -1},
	}
	if got := findCheckScore(checks, "CI-Tests"); got != 10 {
		t.Errorf("expected 10, got %d", got)
	}
	if got := findCheckScore(checks, "Branch-Protection"); got != -1 {
		t.Errorf("expected -1, got %d", got)
	}
	if got := findCheckScore(checks, "Missing"); got != -1 {
		t.Errorf("expected -1 for missing, got %d", got)
	}
	if got := findCheckScore(nil, "CI-Tests"); got != -1 {
		t.Errorf("expected -1 for nil slice, got %d", got)
	}
}

func TestWeakChecks(t *testing.T) {
	checks := []scorecardCheck{
		{Name: "CI-Tests", Score: 10},
		{Name: "SAST", Score: 3},
		{Name: "Branch-Protection", Score: -1},
		{Name: "Code-Review", Score: 0},
	}

	// Default security checks filter
	got := weakChecks(checks, nil)
	found := map[string]bool{}
	for _, w := range got {
		found[w] = true
	}
	if found["CI-Tests(10)"] {
		t.Error("CI-Tests with score 10 should not be weak")
	}
	if !found["SAST(3)"] {
		t.Error("SAST with score 3 should be weak")
	}
	if !found["Branch-Protection(-1)"] {
		t.Error("Branch-Protection with score -1 should be weak")
	}

	// Custom filter
	custom := map[string]bool{"CI-Tests": true}
	got2 := weakChecks(checks, custom)
	if len(got2) != 0 {
		t.Errorf("CI-Tests score 10 should not be weak, got %v", got2)
	}

	// Empty checks
	if got3 := weakChecks(nil, nil); len(got3) != 0 {
		t.Errorf("nil checks should return empty, got %v", got3)
	}
}
