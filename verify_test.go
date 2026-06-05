package main

import (
	"math"
	"strings"
	"testing"
)

func TestCVSSBaseScore(t *testing.T) {
	cases := []struct {
		vec  string
		want float64
	}{
		{"CVSS:3.1/AV:N/AC:L/PR:N/UI:N/S:U/C:H/I:H/A:H", 9.8},
		{"CVSS:3.1/AV:N/AC:L/PR:N/UI:N/S:U/C:N/I:N/A:H", 7.5},
		{"CVSS:3.1/AV:N/AC:L/PR:N/UI:N/S:U/C:L/I:N/A:N", 5.3},
		{"CVSS:3.1/AV:N/AC:H/PR:N/UI:N/S:C/C:H/I:H/A:N", 8.7}, // scope-changed (verified against official calc)
		{"CVSS:3.1/AV:N/AC:L/PR:N/UI:N/S:U/C:N/I:N/A:N", 0.0},
	}
	for _, c := range cases {
		got, ok := cvssBaseScore(c.vec)
		if !ok {
			t.Fatalf("%s: not ok", c.vec)
		}
		if math.Abs(got-c.want) > 0.05 {
			t.Errorf("%s: got %.1f want %.1f", c.vec, got, c.want)
		}
	}
	if _, ok := cvssBaseScore("CVSS:3.1/AV:N/AC:L"); ok {
		t.Error("incomplete vector should not be ok")
	}
}

func TestCVSSBand(t *testing.T) {
	cases := map[float64]string{0: "None", 3.9: "Low", 4.0: "Medium", 6.9: "Medium", 7.0: "High", 8.7: "High", 8.9: "High", 9.0: "Critical", 10.0: "Critical"}
	for s, want := range cases {
		if got := cvssBand(s); got != want {
			t.Errorf("band(%.1f)=%s want %s", s, got, want)
		}
	}
}

func TestVerifyReport(t *testing.T) {
	ctx := &auditContext{}
	// authoritative pin + a real commit in evidence
	ctx.CICD.PinnedSuggestions = "actions/checkout@v4 -> 11bd71901bbe5b1630ceea73d27597364c9af683 | .github/workflows/ci.yml:10"
	ctx.Git.RecentCommits = "abc123 fix things (#42)\n./src/app.ts:99:  const x = 1"

	report := strings.Join([]string{
		"uses: actions/checkout@v4  # pin: 11bd71901bbe5b1630ceea73d27597364c9af683", // authoritative -> verified
		"pin to deadbeefdeadbeefdeadbeefdeadbeefdeadbeef for safety",                 // fabricated -> unverified
		"CVSS 8.7 (Critical)",              // band wrong -> unverified
		"CVSS 9.8 (Critical)",              // band ok
		"see ./src/app.ts:99 for the call", // file:line in evidence -> verified
		"also files.ts:285 is vulnerable",  // file:line NOT in evidence -> unverified
		"fixed in #42",                     // PR in evidence -> verified
		"introduced by #9999",              // PR not in evidence -> unverified
		"vector CVSS:3.1/AV:N/AC:L/PR:N/UI:N/S:U/C:L/I:N/A:N here",
	}, "\n")

	out := verifyReport(report, ctx)
	must := []string{
		"Appendix: Claim Verification",
		"DRAFT",
		"band for 8.7 is High",
		"likely fabricated",
		"no matching line in collected evidence",
		"computed 5.3 (Medium)",
	}
	for _, m := range must {
		if !strings.Contains(out, m) {
			t.Errorf("output missing %q", m)
		}
	}
	// authoritative pin SHA must be verified, not fabricated-flagged
	if strings.Contains(out, "11bd71901bbe5b1630ceea73d27597364c9af683 | SHA | ❌") {
		t.Error("authoritative pin SHA wrongly flagged")
	}
	t.Logf("\n%s", out[strings.Index(out, "## Appendix"):])
}

func TestResolvePinnedActionsLive(t *testing.T) {
	in := ".github/workflows/ci.yml:10:      uses: actions/checkout@v4\n" +
		".github/workflows/ci.yml:11:      uses: actions/checkout@v4\n" +
		".github/workflows/rel.yml:5:      uses: actions/setup-node@v4"
	out := resolvePinnedActions(in)
	if !strings.Contains(out, "actions/checkout@v4 ->") || !strings.Contains(out, "actions/setup-node@v4 ->") {
		t.Fatalf("missing entries:\n%s", out)
	}
	// dedupe: checkout appears twice in input, once in output
	if strings.Count(out, "actions/checkout@v4 ->") != 1 {
		t.Errorf("checkout not deduped:\n%s", out)
	}
	resolved := reSHA40.FindAllString(out, -1)
	t.Logf("resolved %d SHA(s) live:\n%s", len(resolved), out)
}
