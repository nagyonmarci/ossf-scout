package main

import "testing"

// Tools that parse untrusted repos must not see our secrets.
func TestShInEnvScrubsSecrets(t *testing.T) {
	t.Setenv("ANTHROPIC_API_KEY", "s3cret")
	t.Setenv("GITHUB_TOKEN", "s3cret")
	got := shIn(t.TempDir(), "fail", `printf '%s|%s' "${ANTHROPIC_API_KEY:-none}" "${GITHUB_TOKEN:-none}"`)
	if got != "none|none" {
		t.Fatalf("secrets leaked to child env: %q", got)
	}
}

// Untrusted values passed via env must reach the script as data, not code.
func TestShInEnvNoInjection(t *testing.T) {
	payload := `x"; echo INJECTED; echo "`
	got := shIn(t.TempDir(), "fail", `printf '%s' "$V"`, "V="+payload)
	if got != payload {
		t.Fatalf("got %q, want %q", got, payload)
	}
}
