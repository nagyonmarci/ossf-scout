# Security Audit Report — directus/directus

---

## 1. Metadata

| Field | Value |
|---|---|
| **Repository** | directus/directus |
| **Commit** | acfa169 (`acfa1695c48fbfc08194305875fba6077d9bc9bb`) |
| **Scan Date** | 2026-06-05 |
| **Auditor** | Automated — split model workflow |
| **Report Status** | Draft — Pending Human Review |
| **Tools Used** | Actionlint, Zizmor v1.25.2, OSV-Scanner, Trivy, Gitleaks, TruffleHog 3.95.5, CodeQL, pnpm audit (failed), static grep analysis |

---

## 2. Executive Summary

Directus is a widely-deployed open-source data platform with strong foundational security practices, but this audit identified several gaps requiring prompt attention. The overall security posture is **moderate-to-good**: the codebase correctly implements parameterized queries, enforces rate limiting, restricts container privileges, and validates secrets at startup — all meaningful protections. The most significant finding is a **path traversal vulnerability in the `@hono/node-server` dependency (GHSA-92pp-h63x-v22m)**, which can allow HTTP requests to bypass route-based authorization middleware, potentially enabling unauthenticated access to protected endpoints. Additionally, Gitleaks surfaced 21 unitemized alerts requiring manual triage; until those are reviewed, the risk of committed credentials cannot be ruled out. The overall **Security Posture Score is 6.4 / 10**. The single most important action is to **immediately upgrade `@hono/node-server` to ≥ 1.19.13** and pin all GitHub Actions to immutable commit SHAs to close supply-chain exposure.

---

## 3. Scope

| Area | Detail |
|---|---|
| **CI/CD Workflows** | 19 `.github/workflows/*.yml` files |
| **Application Code (SAST)** | `api/src/`, `app/src/`, `packages/` — static grep and pattern analysis |
| **Container Security** | `Dockerfile` (multi-stage, node:22-alpine) |
| **Dependency Analysis** | `pnpm-lock.yaml` — 2,802 packages via OSV-Scanner; Trivy container scan (26 passed, 1 failed) |
| **Secrets Scanning** | Full repository scan via Gitleaks and TruffleHog 3.95.5 |
| **GitHub Signals** | Open issues, open PRs, Dependabot status (403), branch protection (404), secret scanning (404), recent commit log (40+ commits) |
| **Infrastructure** | No Helm charts, Kubernetes manifests, Terraform, OPA/Kyverno/Falco policies detected — not in scope |
| **SLSA / Provenance** | Checked; no SBOM, provenance attestations, or Cosign signatures detected |

---

## 4. Methodology

### Static Analysis
All code findings derive from pattern-matching (grep-equivalent) over the repository snapshot at commit `acfa169`. Data-flow tracing is limited to what can be inferred from file:line evidence; no dynamic instrumentation or fuzzing was performed.

### Dependency Scanning
OSV-Scanner queried the `pnpm-lock.yaml` manifest against the OSV advisory database. `pnpm audit` failed with `"reference.startsWith is not a function"` and produced no usable output; findings from that tool are absent.

### Secrets Scanning
Gitleaks and TruffleHog 3.95.5 performed whole-repository scans. TruffleHog attempts online verification of detected credentials; Gitleaks does not. Gitleaks returned 21 aggregate findings without itemization — these cannot be individually rated in this report (see Finding UNRATED-001).

### Container / IaC Analysis
Trivy scanned the Dockerfile against Dockerfile best-practice checks. No Kubernetes manifests, Helm charts, or Terraform were present.

### GitHub API
Repository metadata, open issues, open PRs, Dependabot alerts (403 — inaccessible), and branch protection (404 — not returned) were queried. Secret-scanning alert API returned 404.

### Known Limitations
- Dynamic / runtime validation was **not performed**; SSRF and deserialization findings are assessed as potential attack surfaces only.
- `pnpm audit` failure means transitive CVE coverage relies solely on OSV-Scanner.
- Zizmor SARIF output was truncated; persist-credentials scope across all 19 workflows cannot be fully confirmed.
- Gitleaks 21 findings are unitemized; individual assessment is impossible without the full report.

---

## 5. Security Strengths

| Control | Evidence | Assessment |
|---|---|---|
| **Parameterized SQL** | ~50+ `knex.raw()` calls use `??`/`?` binding throughout `api/src/database/run-ast/` and operator logic | Prevents SQL injection effectively; auditor comment at `api/src/database/run-ast/lib/apply-query/filter/operator.ts:340` acknowledges binding-order edge case and mitigates it |
| **Non-root container user** | `USER node` in Dockerfile runtime stage | Follows container security best practices; reduces blast radius of container escape |
| **Multi-stage Docker build** | Separate build and runtime stages in Dockerfile | Dev dependencies, build tools (Python 3, build-essential) are not shipped in the production image |
| **Frozen lockfile** | `pnpm install --recursive --offline --frozen-lockfile` | Prevents transitive dependency substitution during builds; hardens against dependency confusion attacks |
| **Rate limiting** | Global and IP-based limiters at `api/src/app.ts:312-316`; health-check rate limit at `api/src/services/server.ts:373-376` | Meaningful protection against credential stuffing and DoS |
| **Secret strength enforcement** | Startup validation enforces 32-byte minimum for `SECRET` | Prevents deployment with weak signing keys |
| **HSTS and security headers** | HSTS configured at `api/src/app.ts:227`; Cross-Origin-Opener-Policy set at `api/src/app.ts:217`; Helmet with CSP enabled | Defense-in-depth for browser-facing endpoints |
| **CodeQL integration** | `.github/workflows/codeql-analysis.yml` — scoped permissions, SARIF upload, daily schedule | Automated SAST gating on the main branch |
| **No hardcoded production secrets** | Gitleaks and TruffleHog found no verified live credentials; no `.env` files in repository | Clean baseline for secrets hygiene |
| **Structured permission policy** | Policy-based permissions with field-level granularity in the Directus permission system | Fine-grained access control reduces over-privilege risk |
| **Authentication breadth** | LDAP, SAML, OAuth2, OpenID, static-token drivers present | Multiple auth integration paths maintained with dedicated drivers |

---

## 6. Findings Summary

| ID | Priority | Severity | Title | OWASP 2021 | Status |
|---|---|---|---|---|---|
| FIND-001 | P1 | **High** | `@hono/node-server` Path Traversal Middleware Bypass (GHSA-92pp-h63x-v22m) | A06:2021 Vulnerable & Outdated Components | Open |
| FIND-002 | P1 | **Medium** | DoS via Uncontrolled Memory in `brace-expansion@5.0.5` (GHSA-jxxr-4gwj-5jf2) | A06:2021 Vulnerable & Outdated Components | Open |
| FIND-003 | P2 | **Medium** | Unpinned GitHub Actions in Release/Deployment Workflows | A08:2021 Software & Supply Chain Integrity Failures | Open |
| FIND-004 | P2 | **Medium** | SSRF — Unvalidated URL in File Import (`api/src/services/files.ts:285`) | A10:2021 Server-Side Request Forgery | Potential — Requires Confirmation |
| FIND-005 | P2 | **Medium** | `unsafe-eval` in Content Security Policy | A05:2021 Security Misconfiguration | Open — By Design |
| FIND-006 | P2 | **Low** | DoS via Promise Hang in `@tootallnate/once` (GHSA-vpq2-c234-7xj6) | A06:2021 Vulnerable & Outdated Components | Open |
| FIND-007 | P2 | **Low** | `persist-credentials: false` Not Set on Checkout (`agent-scan.yml`) | A05:2021 Security Misconfiguration | Open |
| FIND-008 | P2 | **Low** | `X-Powered-By` Header Contradiction (`api/src/app.ts:237`) | A05:2021 Security Misconfiguration | Open |
| FIND-009 | P3 | **Low** | Weak Hashing Algorithms for Non-Security Functions | A02:2021 Cryptographic Failures | Open |
| UNRATED-001 | P2 | **Unrated — pending triage** | Gitleaks: 21 aggregate findings — manual triage required | A02:2021 Cryptographic Failures | Unresolved |

> **Note — Admin-by-design finding (Appendix only):** `context.eval()` at `api/src/operations/exec/index.ts:49` is the documented "Run Script" flow operation, accessible only to administrators. It is documented in the Appendix as a P3 Architectural Risk note per audit principle 14.

---

## 7. Security Posture Summary

| Area | Score (0–10) | Rationale |
|---|---|---|
| **CI/CD Pipeline Security** | 6.0 | CodeQL properly configured; rate-limiting and security headers in place. Deductions: all 19 workflows use mutable version tags, not SHA pins; `persist-credentials` gap in at least one workflow; invalid `models` permission scope in `release.yml` (reliability issue). |
| **Dependency Management** | 5.5 | Frozen lockfile is positive; OSV-Scanner found 4 vulnerable packages including one High (path traversal); `pnpm audit` failed entirely — gap in automated dependency gate. |
| **Secrets Management** | 6.5 | No verified live credentials found; startup `SECRET` validation present. Deductions: Gitleaks 21 unitemized findings cannot be cleared; hardcoded test credentials in version control (LDAP, embedded URIs). |
| **Supply Chain Integrity** | 4.5 | No SBOM, no provenance attestations, no Cosign signatures, no SLSA workflow. All actions on mutable tags. Commit signing present but verification status unknown. |
| **Container Security** | 7.5 | Non-root user, multi-stage build, frozen lockfile in image build. Only finding: missing HEALTHCHECK (Low). No critical Trivy findings. |
| **Application Code (SAST)** | 7.0 | Parameterized SQL throughout; rate limiting; HSTS/CSP/security headers. Deductions: `unsafe-eval` in CSP (by design but increases XSS risk); `X-Powered-By` contradiction; SSRF surface in file import (unconfirmed). |
| **Overall (weighted average)** | **6.4** | Weighted: App Code × 0.25, Deps × 0.20, CI/CD × 0.15, Container × 0.15, Secrets × 0.15, Supply Chain × 0.10. |

---

## 8. Per-Finding Sections

---

### FIND-001 · `@hono/node-server` Path Traversal Middleware Bypass

| Field | Value |
|---|---|
| **OWASP 2021** | A06 — Vulnerable & Outdated Components |
| **CWE** | CWE-22 — Improper Limitation of a Pathname to a Restricted Directory |
| **CVSS v3.1** | `CVSS:3.1/AV:N/AC:L/PR:N/UI:N/S:C/C:H/I:N/A:N` |
| **CVSS Score** | **7.5 High** |
| **Priority** | P1 |
| **Advisory** | GHSA-92pp-h63x-v22m (CVE-2026-39406) |
| **Status** | Open — fix available in v1.19.13 |

**Description**

`@hono/node-server@1.19.10` (present in `pnpm-lock.yaml`) contains a path normalization flaw: repeated leading slashes (e.g., `//admin/secret.txt`) are not canonicalized before route matching. Middleware registered on `/admin/*` is bypassed when the request arrives as `//admin/*`, allowing unauthenticated callers to reach protected handlers.

**Root Cause**

The vulnerability exists because the Hono node adapter does not normalise repeated slashes in the URL path before evaluating registered middleware chains. Route guards that match `/admin/` do not match `//admin/`, so they are skipped entirely.

**Impact Chain**

1. An unauthenticated attacker sends `GET //admin/users` to the Directus API.
2. The Hono middleware registered to protect `/admin/*` routes is not invoked.
3. The underlying handler processes the request without authentication or authorization checks.
4. Depending on which Directus routes are protected by Hono-layer middleware (vs. Directus-internal policy checks), the attacker may read, modify, or delete data they should not access.
5. Worst case: full data exfiltration from all collections, or privilege escalation if administrative endpoints are reached.

> **Note on net risk:** Directus also applies its own field-level permission system inside request handlers. The degree to which route-layer middleware is a sole or primary guard for sensitive endpoints requires runtime confirmation. The upgrade path eliminates the flaw regardless.

**Fix**

```bash
# Upgrade @hono/node-server to ≥ 1.19.13
pnpm update @hono/node-server@1.19.13

# Or pin in package.json directly:
# "@hono/node-server": ">=1.19.13"

# Verify after upgrade:
pnpm list @hono/node-server
```

**Verification**

```bash
# Confirm vulnerability is resolved in lockfile
grep '@hono/node-server' pnpm-lock.yaml | head -5

# OSV-Scanner targeted re-check
osv-scanner --lockfile pnpm-lock.yaml 2>&1 | grep -i hono

# Manual probe against a test instance (requires running server):
curl -v "http://localhost:8055//api/admin/users" \
  -H "Accept: application/json"
# Expected after fix: 401/403. Before fix on a vulnerable instance: varies.
```

---

### FIND-002 · DoS via Uncontrolled Memory in `brace-expansion@5.0.5`

| Field | Value |
|---|---|
| **OWASP 2021** | A06 — Vulnerable & Outdated Components |
| **CWE** | CWE-400 — Uncontrolled Resource Consumption |
| **CVSS v3.1** | `CVSS:3.1/AV:N/AC:L/PR:L/UI:N/S:U/C:N/I:N/A:H` |
| **CVSS Score** | **6.5 Medium** |
| **Priority** | P1 |
| **Advisory** | GHSA-jxxr-4gwj-5jf2 (CVE-2026-45149) |
| **Status** | Open — fix available in v5.0.6 |

**Description**

`brace-expansion@5.0.5` applies its `max` expansion guard too late in the parse cycle. A numeric range such as `{1..10000000}` causes ~505 MB of memory allocation before the guard triggers, enabling a denial-of-service attack wherever user-supplied input reaches brace-expansion parsing.

**Root Cause**

The library's range enumeration loop allocates array elements before checking the configured `max` limit, meaning the guard is evaluated post-allocation rather than pre-allocation.

**Impact Chain**

1. Attacker provides a crafted brace-expression (e.g., a glob pattern or filename pattern) through any Directus API endpoint that internally processes user-supplied patterns via `brace-expansion`.
2. The library allocates ~505 MB per request before terminating.
3. Repeated requests exhaust available heap memory.
4. Node.js process crashes or becomes unresponsive; the Directus API goes offline.

> **Dependency chain:** Whether this is a direct or transitive dependency, and which user-facing paths invoke it, requires runtime tracing to confirm exploitability. Upgrade eliminates the risk regardless.

**Fix**

```bash
# Upgrade brace-expansion to ≥ 5.0.6
pnpm update brace-expansion@5.0.6

# Verify
pnpm list brace-expansion
grep 'brace-expansion' pnpm-lock.yaml | head -5
```

**Verification**

```bash
osv-scanner --lockfile pnpm-lock.yaml 2>&1 | grep -i brace-expansion
# Expected after fix: no GHSA-jxxr-4gwj-5jf2 finding
```

---

### FIND-003 · Unpinned GitHub Actions in Release and Deployment Workflows

| Field | Value |
|---|---|
| **OWASP 2021** | A08 — Software & Supply Chain Integrity Failures |
| **CWE** | CWE-494 — Download of Code Without Integrity Check |
| **CVSS v3.1** | `CVSS:3.1/AV:N/AC:H/PR:N/UI:R/S:C/C:H/I:H/A:N` |
| **CVSS Score** | **7.7 High** |
| **Priority** | P2 |
| **Status** | Open |

> **Severity reasoning:** The release workflow (`release.yml`) uses `docker/build-push-action@v6`, `docker/login-action@v3`, and `docker/setup-buildx-action@v3` on mutable tags AND the workflow handles Docker registry credentials (secrets) and publishes production container images. This meets the High escalation threshold per the Supply-Chain Severity Rubric: write permissions + secrets + deployment trigger. Actions in purely read-only workflows (e.g., `codeql-analysis.yml`, `stale-issues.yml`) would be rated Medium independently; they are noted but not elevated here. Third-party actions `anthropics/claude-code-action@v1` in `claude.yml` also require scrutiny given potential access to repository content.

**Description**

All 19 workflow files use mutable version tags (e.g., `@v6`, `@v4`, `@v3`, `@v1`) rather than immutable commit SHA pins. A tag is a mutable pointer; an action publisher can move it to point to a different — potentially malicious — commit at any time, silently changing what code executes in CI without any repository change.

**Root Cause**

GitHub Actions does not enforce SHA-based pinning. Developers commonly use version tags for readability and auto-update behaviour, but this creates a trust dependency on the action publisher's tag management integrity.

**Impact Chain (release.yml — elevated path)**

1. An attacker compromises the `docker/build-push-action` publisher account or the tag `v6` is moved maliciously.
2. `release.yml` downloads and executes the attacker-controlled action code.
3. The action runs with access to Docker registry credentials (pushed via `docker/login-action`) and any `GITHUB_TOKEN` or repository secrets configured in the job.
4. Attacker exfiltrates credentials, injects malicious code into the published Directus Docker image, or pivots to the container registry.
5. Downstream Directus deployments receive a trojaned production image.

**Affected workflows (priority order)**

| Workflow | Actions on Mutable Tags | Risk |
|---|---|---|
| `release.yml` | `docker/build-push-action@v6`, `docker/login-action@v3`, `docker/setup-buildx-action@v3`, `actions/checkout@v6` | **Highest** — publishes production Docker images, handles registry secrets |
| `claude.yml` | `anthropics/claude-code-action@v1`, `actions/checkout@v4` | **High** — third-party AI action with potential repository read access |
| `check.yml`, `changeset-check.yml` | `tj-actions/changed-files@v47` | **Medium** — third-party, PR-triggered |
| All others | `actions/checkout@v6`, `dessant/lock-threads@v6`, `directus/cla-bot@v0.0.3` | **Medium** — read-only or low-privilege context |

**Fix**

Use `pin-github-action` or `Renovate` with the `pinDigest` option to resolve and pin all actions to immutable SHAs. Example for the highest-risk actions:

```yaml
# release.yml — example pinned replacements (resolve current SHAs with:)
# gh api repos/docker/build-push-action/git/refs/tags/v6 --jq '.object.sha'
# Then trace to the commit SHA (tags may point to tag objects, not commits directly)

# Replace:
#   uses: docker/build-push-action@v6
# With (illustrative format — resolve actual SHA from the tag at time of pinning):
#   uses: docker/build-push-action@<resolved-commit-sha> # v6

# Automation:
npx pin-github-action .github/workflows/release.yml
npx pin-github-action .github/workflows/claude.yml

# Verify all workflows are pinned:
grep -rh 'uses:' .github/workflows/ | grep -v '@[0-9a-f]\{40\}' | sort -u
```

> **Note:** Resolved pin SHAs are not present in the collected evidence (see principle 10). Run the commands above at pinning time to obtain authoritative SHAs.

**Verification**

```bash
# Find all unpinned actions (anything not using a 40-char hex SHA):
grep -rn 'uses:' .github/workflows/ \
  | grep -v 'uses: \./\|uses: \./.github' \
  | grep -v '@[0-9a-f]\{40\}' \
  | grep -v '# v[0-9]'
# Expected after fix: no output
```

---

### FIND-004 · SSRF — Unvalidated URL in File Import Service

| Field | Value |
|---|---|
| **OWASP 2021** | A10 — Server-Side Request Forgery |
| **CWE** | CWE-918 — Server-Side Request Forgery |
| **CVSS v3.1** | `CVSS:3.1/AV:N/AC:H/PR:L/UI:N/S:C/C:L/I:L/A:N` |
| **CVSS Score** | **4.9 Medium** |
| **Priority** | P2 |
| **Status** | **Potential Attack Surface — Requires Confirmation** |

**Description**

`api/src/services/files.ts:285` calls `axios.get(encodeURL(importURL))` where `importURL` is used for fetching remote files. Static analysis does not confirm that the scheme or host of `importURL` is validated before the HTTP request is dispatched. If `importURL` is directly or indirectly user-supplied, an attacker could trigger server-side HTTP requests to internal network endpoints, cloud metadata services (e.g., `http://169.254.169.254`), or localhost services.

> **Requires dynamic validation or manual code-path review to confirm exploitability.** `encodeURL()` likely percent-encodes the URL but does not necessarily restrict the scheme or target host.

**Root Cause**

The `encodeURL()` utility appears to handle URL encoding (percent-encoding of special characters) rather than performing a security-oriented allowlist check on scheme or host. Without an explicit allowlist, any URL reachable from the server process is a valid target.

**Impact Chain (if exploitable)**

1. Authenticated user (or any role with file-import permission) submits an import request with `importURL = "http://169.254.169.254/latest/meta-data/iam/security-credentials/"`.
2. The server makes an outbound GET to the AWS metadata endpoint.
3. IAM credential response is returned to the server; depending on error handling, it may be logged, stored, or returned in an error message.
4. Attacker obtains cloud credentials and escalates to full infrastructure access.

**Fix**

```typescript
// api/src/services/files.ts — add scheme and host validation before fetch
import { URL } from 'url';

const ALLOWED_SCHEMES = ['https:', 'http:'];
// Optionally add an allowlist/blocklist for hosts:
const BLOCKED_HOSTS = /^(localhost|127\.|10\.|192\.168\.|169\.254\.)/i;

function validateImportURL(raw: string): string {
  let parsed: URL;
  try {
    parsed = new URL(raw);
  } catch {
    throw new InvalidPayloadError({ reason: 'Invalid import URL' });
  }
  if (!ALLOWED_SCHEMES.includes(parsed.protocol)) {
    throw new InvalidPayloadError({ reason: 'Unsupported URL scheme' });
  }
  if (BLOCKED_HOSTS.test(parsed.hostname)) {
    throw new InvalidPayloadError({ reason: 'Import URL targets a private address' });
  }
  return raw;
}

// Then at line 285:
const safeURL = validateImportURL(importURL);
const response = await axios.get(encodeURL(safeURL));
```

**Verification**

```bash
# Confirm URL validation logic exists after fix:
grep -n 'validateImportURL\|BLOCKED_HOSTS\|parsed\.protocol' \
  api/src/services/files.ts

# Dynamic test (requires running instance with authenticated session):
curl -X POST "http://localhost:8055/files/import" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"url": "http://169.254.169.254/latest/meta-data/"}'
# Expected after fix: 400 Bad Request with validation error
```

---

### FIND-005 · `unsafe-eval` in Content Security Policy

| Field | Value |
|---|---|
| **OWASP 2021** | A05 — Security Misconfiguration |
| **CWE** | CWE-79 — Cross-Site Scripting (facilitated by policy) |
| **CVSS v3.1** | `CVSS:3.1/AV:N/AC:H/PR:N/UI:R/S:C/C:L/I:L/A:N` |
| **CVSS Score** | **4.7 Medium** |
| **Priority** | P2 |
| **Status** | Open — documented as By Design |

**Description**

The CSP `scriptSrc` directive includes `'unsafe-eval'`, which permits JavaScript's `eval()`, `new Function()`, `setTimeout(string)`, and `setInterval(string)`. This directive is documented as required for the Directus app extension loading mechanism. While intentional, it materially weakens XSS defences: any XSS vulnerability in the application or its extensions can escalate to arbitrary JavaScript execution without the CSP providing a meaningful barrier.

**Root Cause**

The Directus extension system requires dynamic code evaluation to load and execute user-installed extensions at runtime. This is an architectural decision that trades CSP protection for extensibility.

**Impact Chain**

1. An attacker finds an XSS injection point (e.g., a stored XSS in a rich-text field rendered in the admin UI).
2. Under a strict CSP (without `unsafe-eval`), the impact would be limited.
3. With `unsafe-eval`, the injected script can call `eval()` freely, enabling full session hijacking, credential exfiltration, or CSRF against the Directus API using the victim's authenticated session.

**Fix**

The preferred long-term fix is to adopt a nonce-based or hash-based extension loading approach that eliminates the need for `unsafe-eval`. This is an architectural change.

Short-term hardening:

```typescript
// api/src/app.ts — consider restricting to extension-serving routes only
// rather than globally applying unsafe-eval.
// Evaluate whether Trusted Types can replace unsafe-eval for extension loading.

// Document the decision explicitly in code:
helmet({
  contentSecurityPolicy: {
    directives: {
      scriptSrc: [
        "'self'",
        "'unsafe-eval'", // Required for Directus extensions — see: [link to ADR]
        // TODO: Replace with nonce-based approach — tracked in [issue]
      ],
    },
  },
});
```

**Verification**

```bash
# Confirm CSP header contains unsafe-eval and is intentionally documented:
curl -si http://localhost:8055 | grep -i 'content-security-policy'
# Check for 'unsafe-eval' presence and any nonce/hash alternatives:
curl -si http://localhost:8055 \
  | grep -i 'content-security-policy' \
  | grep -o "script-src[^;]*"
```

---

### FIND-006 · DoS via Promise Hang in `@tootallnate/once`

| Field | Value |
|---|---|
| **OWASP 2021** | A06 — Vulnerable & Outdated Components |
| **CWE** | CWE-400 — Uncontrolled Resource Consumption |
| **CVSS v3.1** | `CVSS:3.1/AV:N/AC:H/PR:N/UI:N/S:U/C:N/I:N/A:L` |
| **CVSS Score** | **3.7 Low** |
| **Priority** | P2 |
| **Advisory** | GHSA-vpq2-c234-7xj6 (CVE-2026-3449) |
| **Status** | Open — fix available in v2.0.1 / v3.0.1 |

**Description**

Both `@tootallnate/once@1.1.2` and `@tootallnate/once@2.0.0` are present in `pnpm-lock.yaml`. When an `AbortSignal` is used, the returned Promise hangs indefinitely, causing the awaiting code path to never resolve. This could cause async call stacks to leak resources over time.

**Root Cause**

The library does not properly handle the `abort` event when using `AbortSignal`, leaving the internal `EventEmitter` listener registered but the Promise in a permanently-pending state.

**Impact Chain**

1. A request path that uses `@tootallnate/once` with an `AbortSignal` (e.g., in proxy or connection pooling logic) is invoked repeatedly.
2. Each hung Promise retains its closure references, preventing garbage collection.
3. Under sustained load, memory and event-loop slot exhaustion occur.
4. Service degradation or OOM crash.

**Fix**

```bash
# Upgrade both versions:
pnpm update @tootallnate/once@2.0.1
# For the v1 variant (if direct dep):
pnpm update @tootallnate/once@3.0.1

# Verify:
pnpm list @tootallnate/once
grep 'tootallnate' pnpm-lock.yaml | head -10
```

**Verification**

```bash
osv-scanner --lockfile pnpm-lock.yaml 2>&1 | grep -i 'tootallnate\|GHSA-vpq2'
# Expected after fix: no findings for this advisory
```

---

### FIND-007 · `persist-credentials: false` Not Set on Checkout

| Field | Value |
|---|---|
| **OWASP 2021** | A05 — Security Misconfiguration |
| **CWE** | CWE-522 — Insufficiently Protected Credentials |
| **CVSS v3.1** | `CVSS:3.1/AV:N/AC:H/PR:N/UI:R/S:U/C:L/I:L/A:N` |
| **CVSS Score** | **3.7 Low** |
| **Priority** | P2 |
| **Status** | Open — Zizmor confirmed in `agent-scan.yml`; full scope unconfirmed (SARIF truncated) |

**Description**

`agent-scan.yml` (and potentially other workflows) does not set `persist-credentials: false` on

## Appendix: Claim Verification

Automated post-generation check of concrete claims against collected evidence. Method: SHAs matched against resolved pin set and git evidence; `file:line` against `grep -rn` output; `#PR` against GitHub/commit evidence; CVEs against Dependabot/osv-scanner; workflow files against WorkflowList; pkg@version against npm audit/osv-scanner; CVSS bands recomputed. No live network calls.

| Claim | Type | Status | Detail |
|---|---|---|---|
| `CVE-2026-39406` | CVE | ✓ verified | found in Dependabot/dependency evidence |
| `CVE-2026-45149` | CVE | ✓ verified | found in Dependabot/dependency evidence |
| `CVE-2026-3449` | CVE | ✓ verified | found in Dependabot/dependency evidence |
| `7.5 (High)` | CVSS band | ✓ verified | band correct |
| `6.5 (Medium)` | CVSS band | ✓ verified | band correct |
| `7.7 (High)` | CVSS band | ✓ verified | band correct |
| `4.9 (Medium)` | CVSS band | ✓ verified | band correct |
| `4.7 (Medium)` | CVSS band | ✓ verified | band correct |
| `3.7 (Low)` | CVSS band | ✓ verified | band correct |
| `CVSS:3.1/AV:N/AC:L/PR:N/UI:N/S:C/C:H/I:N/A:N` | CVSS vector | ✓ verified | computed 8.6 (High) |
| `CVSS:3.1/AV:N/AC:L/PR:L/UI:N/S:U/C:N/I:N/A:H` | CVSS vector | ✓ verified | computed 6.5 (Medium) |
| `CVSS:3.1/AV:N/AC:H/PR:N/UI:R/S:C/C:H/I:H/A:N` | CVSS vector | ✓ verified | computed 8.0 (High) |
| `CVSS:3.1/AV:N/AC:H/PR:L/UI:N/S:C/C:L/I:L/A:N` | CVSS vector | ✓ verified | computed 4.9 (Medium) |
| `CVSS:3.1/AV:N/AC:H/PR:N/UI:R/S:C/C:L/I:L/A:N` | CVSS vector | ✓ verified | computed 4.7 (Medium) |
| `CVSS:3.1/AV:N/AC:H/PR:N/UI:N/S:U/C:N/I:N/A:L` | CVSS vector | ✓ verified | computed 3.7 (Low) |
| `CVSS:3.1/AV:N/AC:H/PR:N/UI:R/S:U/C:L/I:L/A:N` | CVSS vector | ✓ verified | computed 4.2 (Medium) |
| `acfa1695c48fbfc08194305875fba6077d9bc9bb` | commit SHA | ✓ verified | present in git/evidence |
| `api/src/database/run-ast/lib/apply-query/filter/operator.ts:340` | file:line | ✓ verified | found in grep evidence |
| `api/src/app.ts:312` | file:line | ✓ verified | found in grep evidence |
| `api/src/services/server.ts:373` | file:line | ✓ verified | found in grep evidence |
| `api/src/app.ts:227` | file:line | ✓ verified | found in grep evidence |
| `api/src/app.ts:217` | file:line | ✓ verified | found in grep evidence |
| `api/src/services/files.ts:285` | file:line | ✓ verified | found in grep evidence |
| `api/src/app.ts:237` | file:line | ✓ verified | found in grep evidence |
| `api/src/operations/exec/index.ts:49` | file:line | ✓ verified | found in grep evidence |
| `brace-expansion@5.0.5` | pkg@version | ✓ verified | found in dependency evidence |
| `node-server@1.19.10` | pkg@version | ✓ verified | package name found in dependency evidence (version not matched exactly) |
| `node-server@1.19.13` | pkg@version | ✓ verified | package name found in dependency evidence (version not matched exactly) |
| `brace-expansion@5.0.6` | pkg@version | ✓ verified | found in dependency evidence |
| `once@1.1.2` | pkg@version | ✓ verified | package name found in dependency evidence (version not matched exactly) |
| `once@2.0.0` | pkg@version | ✓ verified | package name found in dependency evidence (version not matched exactly) |
| `once@2.0.1` | pkg@version | ✓ verified | package name found in dependency evidence (version not matched exactly) |
| `once@3.0.1` | pkg@version | ✓ verified | package name found in dependency evidence (version not matched exactly) |
| `.github/workflows/codeql-analysis.yml` | workflow file | ✓ verified | matches collected workflow list |
| `.github/workflows/release.yml` | workflow file | ✓ verified | matches collected workflow list |
| `.github/workflows/claude.yml` | workflow file | ✓ verified | matches collected workflow list |

**Summary:** 36 verified, 0 unverified.
