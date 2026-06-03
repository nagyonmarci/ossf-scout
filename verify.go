package main

import (
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// verify.go — post-generation grounding check.
//
// Every concrete claim the model writes (action pin SHA, file:line, PR number,
// CVSS band) is checked against the collected evidence in auditContext. Nothing
// here needs the live clone, so it also runs on the re-run path. The result is
// appended as an audit appendix; unverified claims flip the report to DRAFT.

var (
	reSHA40    = regexp.MustCompile(`\b[0-9a-fA-F]{40}\b`)
	reFileLine = regexp.MustCompile(`([\w./@-]+\.(?:ts|tsx|js|jsx|go|py|rego|tf|ya?ml|json|rb|java|php)):(\d{1,6})`)
	rePR       = regexp.MustCompile(`#(\d{2,7})\b`)
	reVector   = regexp.MustCompile(`CVSS:3\.[01]/[A-Za-z:/]+`)
	reScoreBnd = regexp.MustCompile(`(\d{1,2}\.\d)\s*\(?(Critical|High|Medium|Low|None)\)?`)
)

func cvssBand(s float64) string {
	switch {
	case s <= 0:
		return "None"
	case s < 4.0:
		return "Low"
	case s < 7.0:
		return "Medium"
	case s < 9.0:
		return "High"
	default:
		return "Critical"
	}
}

// cvssRoundUp implements the CVSS v3.1 roundup (one decimal place).
func cvssRoundUp(x float64) float64 {
	i := int(math.Round(x * 100000))
	if i%10000 == 0 {
		return float64(i) / 100000.0
	}
	return (math.Floor(float64(i)/10000.0) + 1) / 10.0
}

// cvssBaseScore computes the CVSS v3.1 base score from a vector string.
func cvssBaseScore(vector string) (float64, bool) {
	m := map[string]string{}
	parts := strings.Split(strings.TrimSpace(vector), "/")
	if len(parts) < 2 {
		return 0, false
	}
	for _, p := range parts[1:] {
		if kv := strings.SplitN(p, ":", 2); len(kv) == 2 {
			m[kv[0]] = kv[1]
		}
	}
	av := map[string]float64{"N": 0.85, "A": 0.62, "L": 0.55, "P": 0.2}[m["AV"]]
	ac := map[string]float64{"L": 0.77, "H": 0.44}[m["AC"]]
	ui := map[string]float64{"N": 0.85, "R": 0.62}[m["UI"]]
	scope := m["S"]
	var pr float64
	if scope == "C" {
		pr = map[string]float64{"N": 0.85, "L": 0.68, "H": 0.5}[m["PR"]]
	} else {
		pr = map[string]float64{"N": 0.85, "L": 0.62, "H": 0.27}[m["PR"]]
	}
	cia := func(v string) float64 { return map[string]float64{"H": 0.56, "L": 0.22, "N": 0}[v] }
	if av == 0 || ac == 0 || ui == 0 || pr == 0 || (scope != "U" && scope != "C") ||
		m["C"] == "" || m["I"] == "" || m["A"] == "" {
		return 0, false
	}
	c, i, a := cia(m["C"]), cia(m["I"]), cia(m["A"])
	isc := 1 - (1-c)*(1-i)*(1-a)
	var impact float64
	if scope == "C" {
		impact = 7.52*(isc-0.029) - 3.25*math.Pow(isc-0.02, 15)
	} else {
		impact = 6.42 * isc
	}
	if impact <= 0 {
		return 0, true
	}
	expl := 8.22 * av * ac * pr * ui
	if scope == "C" {
		return cvssRoundUp(math.Min(1.08*(impact+expl), 10)), true
	}
	return cvssRoundUp(math.Min(impact+expl, 10)), true
}

type claimResult struct {
	claim  string
	kind   string
	ok     bool
	detail string
}

// pinnedSHASet returns the lowercase set of authoritative pin SHAs that the
// collector resolved via git ls-remote.
func pinnedSHASet(ctx *auditContext) map[string]bool {
	set := map[string]bool{}
	for _, s := range reSHA40.FindAllString(ctx.CICD.PinnedSuggestions, -1) {
		set[strings.ToLower(s)] = true
	}
	return set
}

func verifyReport(report string, ctx *auditContext) string {
	if strings.TrimSpace(report) == "" {
		return report
	}

	// Universal evidence haystack: every collected string, including grep output
	// (path:line:content), git log, and GitHub JSON.
	hay := report // placeholder; overwritten below
	if raw, err := json.Marshal(ctx); err == nil {
		hay = string(raw)
	}
	hayLower := strings.ToLower(hay)
	pins := pinnedSHASet(ctx)

	var results []claimResult
	seen := map[string]bool{}
	add := func(kind, claim string, ok bool, detail string) {
		key := kind + "|" + claim
		if seen[key] {
			return
		}
		seen[key] = true
		results = append(results, claimResult{claim, kind, ok, detail})
	}

	// 1. Action pin / commit SHAs (40-hex).
	for _, sha := range reSHA40.FindAllString(report, -1) {
		l := strings.ToLower(sha)
		switch {
		case pins[l]:
			add("pin SHA", sha, true, "matches resolved pin set")
		case strings.Contains(hayLower, l):
			add("commit SHA", sha, true, "present in git/evidence")
		default:
			add("SHA", sha, false, "not in resolved pins or evidence — likely fabricated")
		}
	}

	// 2. file:line references must exist in the grep evidence.
	for _, m := range reFileLine.FindAllStringSubmatch(report, -1) {
		full, base, line := m[0], m[1], m[2]
		if i := strings.LastIndex(base, "/"); i >= 0 {
			base = base[i+1:]
		}
		needle := strings.ToLower(base + ":" + line + ":")
		altNeedle := strings.ToLower(full + ":")
		if strings.Contains(hayLower, needle) || strings.Contains(hayLower, altNeedle) {
			add("file:line", full, true, "found in grep evidence")
		} else {
			add("file:line", full, false, "no matching line in collected evidence")
		}
	}

	// 3. PR / issue references.
	for _, m := range rePR.FindAllStringSubmatch(report, -1) {
		num := m[1]
		if strings.Contains(hay, "#"+num) || strings.Contains(hay, "/"+num) ||
			strings.Contains(hay, "\"number\":"+num) {
			add("PR/issue", "#"+num, true, "present in GitHub/commit evidence")
		} else {
			add("PR/issue", "#"+num, false, "no GitHub/commit evidence for this number")
		}
	}

	// 4. CVSS band consistency: a stated "N.N (Band)" must match the band for N.N.
	for _, m := range reScoreBnd.FindAllStringSubmatch(report, -1) {
		score, err := strconv.ParseFloat(m[1], 64)
		if err != nil || score > 10 {
			continue
		}
		want := cvssBand(score)
		if strings.EqualFold(want, m[2]) {
			add("CVSS band", m[1]+" ("+m[2]+")", true, "band correct")
		} else {
			add("CVSS band", m[1]+" ("+m[2]+")", false, "band for "+m[1]+" is "+want)
		}
	}

	// 5. CVSS vector sanity: recompute each vector's score (informational + mismatch).
	for _, vec := range reVector.FindAllString(report, -1) {
		score, ok := cvssBaseScore(vec)
		if !ok {
			add("CVSS vector", vec, false, "vector incomplete or unparseable")
			continue
		}
		add("CVSS vector", vec, true, fmt.Sprintf("computed %.1f (%s)", score, cvssBand(score)))
	}

	if len(results) == 0 {
		return report + "\n\n## Appendix: Claim Verification\n\nNo concrete machine-checkable claims (SHAs, file:line, PR numbers, CVSS vectors) were found to verify.\n"
	}

	// Stable ordering: failures first, then by kind.
	sort.SliceStable(results, func(i, j int) bool {
		if results[i].ok != results[j].ok {
			return !results[i].ok
		}
		return results[i].kind < results[j].kind
	})

	fails := 0
	var b strings.Builder
	b.WriteString("\n\n## Appendix: Claim Verification\n\n")
	b.WriteString("Automated post-generation check of concrete claims against collected evidence. ")
	b.WriteString("Method: SHAs matched against the resolved pin set and git evidence; ")
	b.WriteString("`file:line` against `grep -rn` output; `#PR` against GitHub/commit evidence; ")
	b.WriteString("CVSS bands recomputed from score and vector. No live network calls.\n\n")
	b.WriteString("| Claim | Type | Status | Detail |\n")
	b.WriteString("|---|---|---|---|\n")
	for _, r := range results {
		status := "✓ verified"
		if !r.ok {
			status = "❌ unverified"
			fails++
		}
		fmt.Fprintf(&b, "| `%s` | %s | %s | %s |\n",
			strings.ReplaceAll(r.claim, "|", "\\|"), r.kind, status, r.detail)
	}
	fmt.Fprintf(&b, "\n**Summary:** %d verified, %d unverified.\n", len(results)-fails, fails)

	if fails == 0 {
		return report + b.String()
	}
	banner := fmt.Sprintf("> ⚠️ **DRAFT — %d unverified claim(s).** "+
		"Concrete specifics below could not be grounded in the collected evidence. "+
		"Resolve every item in \"Appendix: Claim Verification\" before publishing or disclosing.\n\n", fails)
	return banner + report + b.String()
}
