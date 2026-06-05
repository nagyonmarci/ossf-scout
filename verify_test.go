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
		{"CVSS:3.1/AV:N/AC:H/PR:N/UI:N/S:C/C:H/I:H/A:N", 8.7}, // scope-changed (verified vs official calc)
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
	// A real commit SHA + a real PR number live in the collected git evidence.
	ctx.Git.RecentCommits = "11bd71901bbe5b1630ceea73d27597364c9af683 fix things (#42)\n./src/app.ts:99:  const x = 1"

	report := strings.Join([]string{
		"pinned to 11bd71901bbe5b1630ceea73d27597364c9af683",         // SHA in evidence -> verified
		"pin to deadbeefdeadbeefdeadbeefdeadbeefdeadbeef for safety", // SHA not in evidence -> unverified
		"CVSS 8.7 (Critical)",              // band wrong -> unverified
		"CVSS 9.8 (Critical)",              // band ok -> verified
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
		"verify it is real",                         // fabricated SHA flagged
		"no matching line in collected evidence",    // bad file:line flagged
		"no GitHub/commit evidence for this number", // #9999 flagged
		"computed 5.3 (Medium)",                     // CVSS vector recomputed
	}
	for _, m := range must {
		if !strings.Contains(out, m) {
			t.Errorf("output missing %q", m)
		}
	}
}

func TestVerifyReportEmpty(t *testing.T) {
	if out := verifyReport("", &auditContext{}); out != "" {
		t.Errorf("empty report should pass through unchanged, got %q", out)
	}
}

func TestVerifyReportNoClaimsCleanPasses(t *testing.T) {
	// A report with no machine-checkable specifics gets the appendix but no DRAFT banner.
	out := verifyReport("All good. No specifics here.", &auditContext{})
	if strings.Contains(out, "DRAFT") {
		t.Error("a claim-free report should not be marked DRAFT")
	}
	if !strings.Contains(out, "Appendix: Claim Verification") {
		t.Error("appendix should still be appended")
	}
}
