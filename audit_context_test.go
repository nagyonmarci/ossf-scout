package main

import "testing"

// Untrusted values passed via env must reach the script as data, not code.
func TestShInEnvNoInjection(t *testing.T) {
	payload := `x"; echo INJECTED; echo "`
	got := shIn(t.TempDir(), "fail", `printf '%s' "$V"`, "V="+payload)
	if got != payload {
		t.Fatalf("got %q, want %q", got, payload)
	}
}
