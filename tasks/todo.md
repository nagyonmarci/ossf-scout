# Audit validity hardening — todo

Goal: make the AI-generated audit *valid* — every concrete claim (action pin SHA,
file:line, #PR/commit, CVSS band, severity) is grounded in evidence or flagged.

## Plan & status

- [x] A. Ground-truth pin SHAs: resolvePinnedActions()/resolveTagSHA() in collectContext;
      authoritative SHA block in buildContextMarkdown; feeds report + verifier allow-set.
- [x] B1. auditSystemPrompt principle 10 (NO FABRICATION) + 11 (CVSS vector + bands;
      "8.7 is HIGH, not Critical").
- [x] B3. auditSystemPrompt principle 12 — SUPPLY-CHAIN SEVERITY RUBRIC: unpinned action
      = Medium by default; High only with write-perms/secrets AND an attacker-influenced
      trigger (pull_request_target/issue_comment/workflow_run/fork PR); never Critical/P0
      for tag-pinning alone.
- [x] B4. fetchGitHubContext classifies list responses: array => "ok"; {message} =>
      "unavailable: <msg>" (rate-limit) with data nulled out. New IssuesStatus/PRsStatus
      fields. ghListAvailable() (backward-compatible with legacy contexts) gates rendering
      in buildContextMarkdown + split sections. Prompt sections 6 (Issues/PRs) and 8
      (Remediation) made conditional in both prompt builders — no specific numbers when
      data is unavailable; cite a fix ref only if present in collected evidence.
- [x] C. verify.go: CVSS v3.1 calculator + post-gen claim verification (SHAs, file:line,
      #PR, CVSS bands/vectors) against auditContext evidence only; "Appendix: Claim
      Verification" + DRAFT banner; wired into runAuditGenerate.
- [x] V. gofmt-clean (added code); full audit.go+verify.go type-checks (go vet, stubbed
      db); go test green incl. live SHA resolution and B4 helper behavior.

## Review

Changed: audit.go (+~213/-20), new verify.go, verify_test.go, audit_github_test.go.
go.mod untouched.

Verification (isolated/stubbed module, Go 1.22, GOTOOLCHAIN=local):
- go vet clean over full audit.go + verify.go (db.go/server.go stubbed; their sqlite/uuid
  deps + frontend embed are sandbox limits, not code issues).
- CVSS calc matches official: 9.8 / 7.5 / 5.3 / 8.7(S:C) / 0.0; incomplete vector rejected.
- Bands correct at boundaries: 8.7 -> High, 9.0 -> Critical.
- verifyReport flags fabricated SHA, wrong band, absent file:line, absent #PR; passes
  authoritative pin SHA, real #PR, present file:line, valid vector; emits DRAFT + counts.
- resolvePinnedActions resolves real SHAs live (actions/checkout@v4, setup-node@v4), deduped.
- ghListAvailable: rate-limited never trusted; legacy empty-status + real array honoured.

Key insight: the original audit's CVSS *score* (8.7) was correct; only the *band* label
(Critical) was wrong, and unpinned actions were over-rated P0. Principle 11 + 12 and the
band recompute target exactly those two error classes.

Follow-ups (optional): run the verifier on the static template report too; surface the
DRAFT/clean state as a column in the audit history UI; teach the verifier to parse a
findings table and cross-check Priority vs the rubric in 12.
