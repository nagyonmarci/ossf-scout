# ossf-scout

[![Go](https://img.shields.io/badge/go-1.25-00acd7?logo=go&logoColor=white)](go.mod)
[![License](https://img.shields.io/badge/license-Unlicense-blue)](LICENSE)
[![OpenSSF Scorecard](https://api.securityscorecards.dev/projects/github.com/nagyonmarci/ossf-scout/badge)](https://securityscorecards.dev/viewer/?uri=github.com/nagyonmarci/ossf-scout)
[![OpenSSF Best Practices](https://www.bestpractices.dev/projects/13066/badge)](https://www.bestpractices.dev/projects/13066)

Discover GitHub repositories where security practices are weakest — and where a well-crafted PR can make the biggest impact.

Searches GitHub for popular repositories, queries the [OpenSSF Scorecard](https://scorecard.dev/) API for each, and surfaces projects missing key security practices — no CI tests, no SAST, no branch protection, etc.

Available as a **CLI tool** or a **web server** with a browser UI, scan history, scheduled audits, remediation tracking, and portfolio views.

Point ossf-scout at **any public repo** and it clones it, runs 30+ security tools, and produces a formal Markdown report — findings with CVSS scores and OWASP/STRIDE mapping, fix commands, and ready-to-paste CI guardrails. Free as a static snapshot, or AI-synthesised via Anthropic, OpenAI, Gemini, or Ollama.

**Sample output for `directus/directus`:** [AI report](examples/audit-directus-directus-2026-06-05_1.md) · [PDF](examples/audit-directus-directus.pdf) · [raw evidence context](examples/context-directus-directus-2026-06-05.md)

---

## Quick Start

### CLI

```bash
export GITHUB_TOKEN=ghp_...
go run . -lang go -min-stars 1000 -max-score 5 -limit 50
```

### Web Server

```bash
go run . -serve
# Open http://localhost:7878
```

### Docker

```bash
GITHUB_TOKEN=ghp_... docker compose up --build
# Open http://localhost:7878
```

Scan history is stored in `./data/ossf-scout.db` and persists across restarts.

### Run an audit

Open the **Audit** tab in the web UI, enter `owner/repo`, choose a provider, and click **Run Audit**. The report appears in the browser when complete (~1–3 min for AI, ~30s for static snapshot).

---

## How it compares

Most OpenSSF Scorecard tooling answers *"how secure is **this** repo?"* or *"how is **my org** trending?"*. ossf-scout asks the inverse: **which popular repositories are weakest right now, so I know where a contribution lands hardest?**

| Tool | Focus | Where ossf-scout differs |
|------|-------|--------------------------|
| [`ossf/scorecard`](https://github.com/ossf/scorecard) | The scoring engine + API | ossf-scout consumes the API and adds a **discovery/search layer** on top |
| [`scorecard-visualizer`](https://github.com/ossf/scorecard-visualizer) | Pretty view of one repo's score | ossf-scout ranks **many** repos by weakness, not one at a time |
| [`scorecard-monitor`](https://github.com/ossf/scorecard-monitor) | Tracking score changes for **your own** orgs | ossf-scout finds **other people's** popular-but-weak repos to contribute to |
| [deps.dev](https://deps.dev) / [Socket](https://socket.dev) | Package/dependency (SCA) security | ossf-scout looks at **repo-level** security posture, not packages |

The niche: **GitHub Search + Scorecard filters to surface popular-but-weak repos**, plus an **AI-powered DevSecOps audit/remediation workspace** — all in one self-contained binary.

---

## Features

### Discovery & Scorecard scanning

- **GitHub Search + OpenSSF Scorecard discovery** — ranks popular repositories whose Scorecard results fall below your threshold
- **Scorecard API + CLI fallback** — queries `api.securityscorecards.dev`; optionally falls back to the local `scorecard` binary for repos not yet indexed
- **Single-repo mode** — score any public repo directly by `owner/repo` without running a GitHub Search query
- **GitHub Trending scanner** — scrapes trending repositories by language/time window and scores them against OpenSSF data
- **Organization scan queue** — lists public repos for a GitHub org, filters forks/archived repos/min stars, and queues selected repos for audit
- **Flexible filters** — language, topic, keyword, pushed-after date, min stars, max Scorecard score, min Maintained score, and highlighted checks

### Web workspace

- **Browser UI** — scan form, scan history, result details, sortable/filterable results, resizable columns, sticky headers
- **Persistent SQLite history** — scans, results, audits, audit contexts, schedules, remediation items, and trend data survive restarts
- **Portfolio dashboard** — aggregates watched repos with latest score, stars, weak checks, audit counts, and score sparklines
- **Remediation board** — extracts findings from audit reports into trackable cards with severity, status, notes, and due dates
- **Schedules** — run recurring audits at configurable intervals; auto-suggested for repos that repeatedly appear weak
- **Notifications** — in-app toast and browser Notification API on scan completion; optional outbound webhook (`NOTIFY_WEBHOOK_URL`)

### DevSecOps audits

- **Static snapshot mode** — runs for free without an AI key; returns the collected evidence as a structured Markdown report
- **AI report generation** — supports Anthropic, OpenAI, Gemini, and local/remote Ollama models
- **Split generation** — a cheaper model summarises evidence sections; a stronger model writes the final report
- **Ground-truth claim verification** — every concrete claim (commit SHA, `file:line`, `#PR`, CVE, `pkg@version`, CVSS vector) is checked against the evidence; unverifiable claims are listed in an appendix and the report is marked **DRAFT**
- **Cost tracking and guardrails** — records tokens, estimates per-model cost, and can reject audits above `MAX_AUDIT_COST_USD`
- **GitHub webhook audits** — signed `pull_request`/`push` webhooks trigger a free static audit when security-sensitive files change

### Packaging

- **Self-contained binary** — React frontend embedded via `//go:embed`, SQLite via pure-Go driver
- **Docker image with bundled tools** — ships `scorecard`, `gitleaks`, `actionlint`, `osv-scanner`, `trivy`, `helm`, `zizmor`, `kube-linter`, `trufflehog`, `semgrep`, Node/npm, and `pnpm`; amd64 + arm64
- **Hardened image** — [Wolfi](https://github.com/wolfi-dev/os) base (near-zero OS CVEs), non-root user (UID 10001), SHA256-verified binary downloads, static Go build, `dumb-init` PID 1, `HEALTHCHECK`, OCI labels
- **Offline-friendly Ollama profile** — Docker Compose can run an Ollama sidecar for local AI generation

---

## Screenshots

**Scan configuration and history**

![Scans](screenshots/scans.png)

**Scan results — repos with weak Scorecard scores**

![Results](screenshots/results.png)

**GitHub Trending scored against Scorecard API**

![Trending](screenshots/trending.png)

---

## DevSecOps Audit

### Providers

| Provider | Cost | Setup |
|----------|------|-------|
| **Static snapshot** | Free | No key needed — returns structured raw evidence |
| **Anthropic** | Paid | API key via UI or `ANTHROPIC_API_KEY`; models: Opus 4.8, Sonnet 4.6, Haiku 4.5 and earlier |
| **OpenAI** | Paid | API key via UI or `OPENAI_API_KEY`; models: GPT-5.5, GPT-4.1 family, GPT-4o family, o-series |
| **Gemini** | Paid | API key via UI or `GEMINI_API_KEY`; models: Gemini 2.5 Pro/Flash, 2.0 Flash, 1.5 Pro/Flash |
| **Ollama** | Free/local | `OLLAMA_BASE_URL` on the server; model name selected in the UI |

Per-run token counts and estimated cost are shown in the audit history. Set `MAX_AUDIT_COST_USD` to reject runs above your budget.

**Split generation:** enable in the UI to have a fast model summarise each evidence section first; the selected model synthesises the final report. Recommended for large monorepos. Implemented for Anthropic and Ollama.

**Ollama setup:**

```bash
ollama serve
ollama pull llama3.2   # or qwen2.5, deepseek-r1:8b, etc.
```

For Docker: the default `http://host.docker.internal:11434` is pre-configured in `docker-compose.yml`. For offline use:

```bash
OLLAMA_BASE_URL=http://ollama:11434 docker compose --profile offline up --build
```

### What it collects

| Category | Checks |
|----------|--------|
| CI/CD | Unpinned GitHub Actions, `zizmor` workflow analysis, `actionlint` workflow linting, workflow file list |
| Code | `eval()`, `Math.random()`, raw SQL, hardcoded secrets, weak crypto, SQL injection (`knex.raw`/`whereRaw`), SSRF (`fetch`/`axios`/`got`) with surrounding context, path traversal, XXE, deserialization, rate limiting, CORS config, Semgrep auto findings |
| Key files | Entry point (first 150 lines), auth middleware, permission system, security config (`helmet`/`cors`/`session`), startup validation, `CODEOWNERS` |
| Infrastructure | Dockerfile(s) with static analysis (missing USER, unpinned FROM, hardcoded secrets in ENV/ARG, dangerous packages), `helm lint`, Helm secret templates + values |
| Dependencies | `pnpm audit` / `npm audit` / `yarn audit` JSON, workspace `overrides` |
| Git history | Last 30 commits, files changed in the last 10 commits |
| GitHub API | Open issues (up to 50), open PRs (up to 20), secret-scanning alerts, branch protection rules, Dependabot alerts, release history |
| Secrets | `gitleaks`, `trufflehog`, private key headers, `.env` file contents, AWS/JWT/GH token regex patterns |
| IaC | Terraform file list, `trivy config`, `osv-scanner`, Kubernetes manifest list, `kube-linter` |
| Policy as Code | OPA `.rego` files, Kyverno `ClusterPolicy`/`Policy` YAMLs, Falco rule detection |
| SLSA / Supply Chain | Provenance / SBOM files, cosign keys, SLSA GitHub Generator workflow usage, signed commit check |

All tools are **bundled in the Docker image** (amd64 + arm64) — no separate installation needed.

### Report structure

The generated Markdown document follows a fixed structure:

- **Executive summary** — 3–5 sentence non-technical overview for managers and CISOs
- **Findings summary** — table: ID · Priority (P0–P3) · CVSS severity · Title · OWASP 2021 · Status
- **Per-finding sections** — Root Cause · Impact Chain · Fix · Verification shell commands
- **Security posture scores** — area scores (0–10) for CI/CD, Dependencies, Secrets, Supply Chain, Container, App Code
- **Remediation roadmap** — 30/60/90-day horizon mapped to Finding IDs; Sprint Backlog with story points
- **Shift-left guardrails** — CI gate per finding with a runnable GitHub Actions YAML snippet
- **Appendix + Threat model** — full STRIDE assessment across auth flows, CI/CD, supply chain, secrets, and API inputs

---

## CLI Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-lang` | _(any)_ | Filter by language (`go`, `python`, `java`, …) |
| `-topic` | _(any)_ | Filter by GitHub topic (`llm`, `kubernetes`, …) |
| `-keyword` | _(any)_ | Filter by keyword in repo name or description |
| `-repo` | _(none)_ | Scan a single specific repo (`owner/repo`), skips GitHub search |
| `-pushed-after` | _(none)_ | Only include repos pushed after this date (`YYYY-MM-DD`) |
| `-min-stars` | `500` | Minimum GitHub star count |
| `-max-score` | `5.0` | Maximum OpenSSF score to include (0–10) |
| `-min-maintained` | `0` | Minimum Maintained check score (0 = disabled) |
| `-limit` | `100` | Max repos to fetch from GitHub search |
| `-workers` | `5` | Concurrent Scorecard API queries |
| `-checks` | _(security set)_ | Comma-separated check names to highlight |
| `-json` | `false` | Output as JSON instead of a table |
| `-token` | `$GITHUB_TOKEN` | GitHub personal access token |
| `-cli-fallback` | `false` | Use local `scorecard` CLI for repos not in the Scorecard database |
| `-serve` | `false` | Start web server mode |
| `-port` | `7878` | Port for web server mode |
| `-db` | `ossf-scout.db` | SQLite database path (server mode) |

---

## Environment Variables

| Variable | Description |
|----------|-------------|
| `GITHUB_TOKEN` | GitHub PAT. Without it, GitHub limits unauthenticated requests to 60/hour. |
| `ANTHROPIC_API_KEY` | Anthropic API key for the Audit tab. Can also be provided per-audit in the web UI. |
| `OPENAI_API_KEY` | OpenAI API key for the Audit tab. Can also be provided per-audit in the web UI. |
| `GEMINI_API_KEY` | Google Gemini API key for the Audit tab. Can also be provided per-audit in the web UI. |
| `OLLAMA_BASE_URL` | Base URL for a local/remote Ollama instance (e.g. `http://localhost:11434`). Defaults to `http://host.docker.internal:11434` in Docker. |
| `MAX_AUDIT_COST_USD` | Optional budget cap. Audits whose pre-flight cost estimate exceeds this value are rejected before any tokens are sent. |
| `GITHUB_WEBHOOK_SECRET` | If set, incoming GitHub webhook payloads are validated against the `X-Hub-Signature-256` header — unsigned requests are rejected. |
| `NOTIFY_WEBHOOK_URL` | A URL to POST a JSON notification to after each audit completes (repo, status, audit ID). |
| `AUTH_REQUIRED` | Set to `true` to enable Authentik forward-auth mode. Requests missing `X-Authentik-Username` return 401; write/admin endpoints check `ossf-write`/`ossf-admin` groups. |

### Token scopes

| Mode | Required scopes |
|------|----------------|
| Default (REST API via `api.securityscorecards.dev`) | None — any valid token raises the rate limit |
| `-cli-fallback` (scorecard CLI) | Classic PAT with **`public_repo`** — needed for the `Branch-Protection` check (GraphQL query) |
| Dependabot alerts collection (audit) | **`security_events`** — needed to read Dependabot vulnerability alerts via the GitHub API |

To create a classic PAT: GitHub → Settings → Developer settings → Personal access tokens → Tokens (classic) → tick **`public_repo`** under _repo_.

---

## Building from Source

Prerequisites: **Go 1.25+**, **Node 22+**

```bash
make build
# Produces ./ossf-scout — a single self-contained binary
```

---

## Architecture

Single Go binary. The web server embeds the React frontend via `//go:embed` — no separate static file serving needed. Scan history is stored in SQLite using [`modernc.org/sqlite`](https://pkg.go.dev/modernc.org/sqlite) (pure Go, no CGo required).

The OpenSSF checks evaluated by default:

`CI-Tests` · `SAST` · `Dependency-Update-Tool` · `Vulnerabilities` · `Pinned-Dependencies` · `Branch-Protection` · `Code-Review` · `Maintained`

---

## Contributing

PRs welcome. Run `make dev` to build locally. The web server embeds the frontend at build time — edit `frontend/src/` and rebuild with `make dev`.

---

## Track Record

ossf-scout exists to surface open-source projects with weak security postures so contributors can improve them — so we ran it against this repo too. Applying its own findings took the Scorecard score from **5.2 → 7.7 / 10** through 11 documented fixes (broken Scorecard CI, branch protection, license, security policy, Dockerfile digest pinning, fuzz tests, CodeQL, and more).

📖 Full running log: [docs/SCORECARD-JOURNEY.md](docs/SCORECARD-JOURNEY.md)

---

## License

[The Unlicense](LICENSE) — public domain
