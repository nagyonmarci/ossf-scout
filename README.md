# ossf-scout

[![Go](https://img.shields.io/badge/go-1.25-00acd7?logo=go&logoColor=white)](go.mod)
[![License](https://img.shields.io/badge/license-Unlicense-blue)](LICENSE)
[![OpenSSF Scorecard](https://api.securityscorecards.dev/projects/github.com/nagyonmarci/ossf-scout/badge)](https://securityscorecards.dev/viewer/?uri=github.com/nagyonmarci/ossf-scout)
[![OpenSSF Best Practices](https://www.bestpractices.dev/projects/13066/badge)](https://www.bestpractices.dev/projects/13066)

Discover GitHub repositories where security practices are weakest — and where a well-crafted PR can make the biggest impact.

Searches GitHub for popular repositories, queries the [OpenSSF Scorecard](https://scorecard.dev/) API for each, and surfaces projects missing key security practices — no CI tests, no SAST, no branch protection, etc.

Available as a **CLI tool** or a **web server** with a browser UI and scan history.

> ### 🔎 Generate a full DevSecOps audit
> Point ossf-scout at **any public repo** and it clones it, runs 30+ security tools, and produces a formal Markdown report — findings with CVSS scores and OWASP/STRIDE mapping, fix commands, and ready-to-paste CI guardrails. Free as a static snapshot, or AI-synthesised via Anthropic/Ollama.
>
> **📄 See real samples** — AI-synthesised report → [`examples/directus-audit-report-ai.md`](examples/directus-audit-report-ai.md) &nbsp;·&nbsp; static snapshot → [`examples/directus-audit-report.md`](examples/directus-audit-report.md) &nbsp;·&nbsp; raw evidence context → [`examples/directus-audit-context.md`](examples/directus-audit-context.md)

---

## How it compares

Most OpenSSF Scorecard tooling answers *"how secure is **this** repo?"* or *"how is **my org** trending?"*. ossf-scout asks the inverse question: **which popular repositories are weakest right now, so I know where a contribution lands hardest?**

| Tool | Focus | Where ossf-scout differs |
|------|-------|--------------------------|
| [`ossf/scorecard`](https://github.com/ossf/scorecard) | The scoring engine + API | ossf-scout consumes the API and adds a **discovery/search layer** on top |
| [`scorecard-visualizer`](https://github.com/ossf/scorecard-visualizer) | Pretty view of one repo's score | ossf-scout ranks **many** repos by weakness, not one at a time |
| [`scorecard-monitor`](https://github.com/ossf/scorecard-monitor) | Tracking score changes for **your own** orgs | ossf-scout finds **other people's** popular-but-weak repos to contribute to |
| [deps.dev](https://deps.dev) / [Socket](https://socket.dev) | Package/dependency (SCA) security | ossf-scout looks at **repo-level** security posture, not packages |

The niche: **GitHub Search + Scorecard filters to surface popular-but-weak repos**, plus an **AI-powered DevSecOps audit/remediation** report — all in one self-contained binary.

---

## Features

- **Scorecard API + CLI fallback** — queries `api.securityscorecards.dev`; falls back to the local `scorecard` binary for repos not yet indexed
- **Web UI** — filterable results table, resizable columns, sticky header, scan history
- **Quick presets** — DevSecOps opportunities, AI/LLM, MCP/Agents, Cloud Native, Security tooling
- **Filters** — language, topic, keyword, min stars, date range, min Maintained score, specific checks
- **Single-repo mode** — score any repo directly by `owner/repo`
- **Open issues count** — from GitHub Search API (no extra token scopes)
- **Notifications** — in-app toast + browser Notification API on scan completion
- **Self-contained binary** — React frontend embedded via `//go:embed`, SQLite via pure-Go driver
- **AI-powered Audit tab** — clones any public GitHub repo, runs static analysis, generates a formal DevSecOps report ([sample](examples/directus-audit-report-ai.md)); choose Anthropic (Opus 4 / Sonnet 4 / Haiku 4), a local **Ollama** model, or run for free as a static data snapshot

---

## Screenshots

**Scan configuration and history**

![Scans](screenshots/scans.png)

**Scan results — repos with weak Scorecard scores**

![Results](screenshots/results.png)

**GitHub Trending scored against Scorecard API**

![Trending](screenshots/trending.png)

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

The scan history is stored in `./data/ossf-scout.db` and persists across restarts.

### Audit Tab

> **📄 Sample output** for `directus/directus`:
> - **AI-synthesised report** → [`examples/directus-audit-report-ai.md`](examples/directus-audit-report-ai.md) — calibrated findings, CVSS vectors, CI guardrails (the full audit format)
> - **Static snapshot** → [`examples/directus-audit-report.md`](examples/directus-audit-report.md) — free, no-AI raw data dump
> - **AI input context** → [`examples/directus-audit-context.md`](examples/directus-audit-context.md) — the evidence fed to the model

Open the **Audit** tab in the web UI:

```bash
go run . -serve
# Navigate to http://localhost:7878 → Audit tab
```

Enter `owner/repo` (e.g. `directus/directus`), choose a provider, and click **Run Audit**:

| Provider | Cost | Setup |
|----------|------|-------|
| **Static snapshot** | Free | No key needed — returns structured raw data |
| **Anthropic** | ~$0.03–$0.50 | API key via UI or `ANTHROPIC_API_KEY` env |
| **Ollama** | Free (local) | Ollama running locally; set URL + model name |

**Anthropic models:**

| Model | Single-stage | Split mode (Haiku analyses sections) |
|-------|-------------|--------------------------------------|
| Opus 4 (most capable) | ~$0.20–$0.50 | ~$0.10–$0.25 |
| Sonnet 4 | ~$0.05–$0.15 | ~$0.03–$0.08 |
| Haiku 4 (fastest / cheapest) | ~$0.01–$0.05 | — |

The tool sends **Markdown context** to the AI instead of raw JSON — the Zizmor SARIF output (which can be 40 000+ lines) is replaced with a compact findings table, reducing input tokens by ~90% compared to the JSON approach. All AI paths (single-stage, split section analysis, split synthesis) use this format.

Enable **Split generation** in the UI to have Haiku summarise each evidence section first; the selected model then synthesises the final report. Recommended for large monorepos where per-section analysis improves finding quality.

**Ollama setup:**

```bash
ollama serve
ollama pull llama3.2   # or qwen2.5, deepseek-r1:8b, etc.
```

In the UI set the Ollama base URL. When running via Docker, leave the URL field empty — the server default (`host.docker.internal:11434`) is pre-configured in `docker-compose.yml`. For native use set it to `http://localhost:11434`. If the model's context window is too small, the tool automatically retries with a compacted context.

The Markdown report appears in the browser when complete (~1–3 min for AI, ~30s for snapshot) and can be downloaded as a `.md` file. If AI generation fails, the static snapshot is saved as a fallback — a **Run with AI** button on the detail page lets you re-run the same repo with a different provider.

A **Download AI context** button is available for every audit (including static snapshots) once the repo has been cloned and analysed. It downloads the compact Markdown that would be sent to the AI — useful for pasting into any LLM manually or for inspecting exactly what evidence was collected.

**What it collects**

| Category | Checks |
|----------|--------|
| CI/CD | Unpinned GitHub Actions, `zizmor` workflow analysis, `actionlint` workflow linting, workflow file list |
| Code | `eval()`, `Math.random()`, raw SQL, `X-Powered-By`, hardcoded secrets, weak crypto, `process.exit`/`os.Exit`, SQL injection (`knex.raw`/`whereRaw`), SSRF (`fetch`/`axios`/`got`), path traversal, XXE, deserialization, rate limiting, CORS config |
| Key files | Entry point (first 150 lines), auth middleware, permission system, security config (`helmet`/`cors`/`session`) |
| Infrastructure | `helm lint`, Helm secret templates + values, `Dockerfile` |
| Dependencies | `pnpm audit` / `npm audit` JSON, workspace `overrides` |
| Git history | Last 30 commits, files changed in the last 10 commits |
| GitHub API | Open issues (up to 50), open PRs (up to 20), secret-scanning alerts, branch protection rules (requires token) |
| Secrets | `gitleaks`, `trufflehog`, private key headers, `.env` file contents, AWS/JWT/GH token regex patterns |
| IaC | Terraform file list, `checkov`, `trivy config`, Kubernetes manifest list, `kube-linter` |
| Policy as Code | OPA `.rego` files, Kyverno `ClusterPolicy`/`Policy` YAMLs, Falco rule detection |
| SLSA / Supply Chain | Provenance / SBOM files, cosign keys, SLSA GitHub Generator workflow usage, signed commit check |

All tools are **bundled in the Docker image** (amd64 + arm64) — no separate installation needed.

**Report structure**

The generated Markdown document follows a fixed structure:

1. Metadata table — date, repo, commit, auditor, status
2. Scope & Methodology
3. Findings Summary — table with ID, Priority (P0–P2), CVSS Severity, OWASP 2021 category
4. Per-finding sections — Root Cause · Impact Chain · Fix · Verification shell commands
5. Open GitHub Issues & PRs — security-relevant items with risk assessment
6. Remediation Status table & Verification Checklist
7. P2 Recommendations — table with Effort (hours/days) and Risk Reduction estimates
8. Shift-left guardrails — maps each finding to an automated CI gate with a runnable GitHub Actions YAML snippet
9. Threat Model (STRIDE) — table covering Spoofing, Tampering, Repudiation, Information Disclosure, DoS, Elevation of Privilege across auth flows, CI/CD, supply chain, secrets, and API inputs
10. Appendix — full assessment across SQL Injection, Auth, Authorisation, SSRF, XXE, Path Traversal, Cryptography, Rate Limiting, Dependencies, HTTP Headers, Container, Kubernetes/Helm

**Cost:** Free without an API key (static snapshot). With a key: Sonnet 4 ~$0.05–$0.15 per run (single-stage). Token counts and estimated cost are shown in the audit history table.

---

## CLI Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-lang` | _(any)_ | Filter by language (`go`, `python`, `java`, …) |
| `-topic` | _(any)_ | Filter by GitHub topic (`llm`, `kubernetes`, …) |
| `-keyword` | _(any)_ | Filter by keyword in repo name or description |
| `-single-repo` | _(none)_ | Score a specific repo directly (`owner/repo`) |
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
| `ANTHROPIC_API_KEY` | Anthropic API key for the Audit tab. Can also be provided per-audit in the web UI (overrides the server-level key). |

The token can also be provided per-scan in the web UI (overrides the server-level token).

### Token scopes

| Mode | Required scopes |
|------|----------------|
| Default (REST API via `api.securityscorecards.dev`) | None — any valid token is enough to raise the rate limit |
| `-cli-fallback` (scorecard CLI) | Classic PAT with **`public_repo`** scope — needed for the `Branch-Protection` check (GraphQL query) |

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

## Eating Our Own Dog Food

ossf-scout exists to surface open-source projects with weak security postures so contributors can improve them — so we ran it against this repo too. Applying its own findings took the Scorecard score from **5.2 → 7.7 / 10** through 11 documented fixes (broken Scorecard CI, branch protection, license, security policy, Dockerfile digest pinning, fuzz tests, CodeQL, and more).

📖 **Full running log:** [docs/SCORECARD-JOURNEY.md](docs/SCORECARD-JOURNEY.md)

---

## License

[The Unlicense](LICENSE) — public domain
