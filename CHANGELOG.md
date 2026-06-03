# Changelog

All notable changes to this project will be documented in this file.
Format based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## [Unreleased]

### Added
- Audit tab — AI-powered DevSecOps report for any public GitHub repo (Claude Opus, static analysis, Markdown output); Anthropic API key configurable server-side or per-audit in the UI

## [0.2.0] - 2026-06-01

### Added
- Go fuzz tests for `parseTrending` and `parseChecks`
- Unit tests (`workers_test.go`, `trending_test.go`) with CI coverage reporting
- `golangci-lint` (errcheck + staticcheck) added to CI
- `CONTRIBUTING.md` — contribution guide
- `SECURITY.md` — vulnerability reporting via GitHub Private Advisory
- OpenSSF Best Practices badge (bestpractices.dev project 13066)
- Branch protection on `main` (no force-push, no deletion, required status check)

### Changed
- Frontend upgraded: React 18 → 19, Vite 6 → 8, TypeScript 5 → 6, @vitejs/plugin-react 4 → 6
- Go dependency: `modernc.org/sqlite` 1.34 → 1.51, `golang.org/x/sys` 0.42 → 0.44
- Dockerfile base images pinned by digest; `scorecard` CLI pinned to v5.5.0

### Fixed
- Sticky header not attaching in Results tab (deferred mount until data loaded)
- Stars, Issues, Score columns wrapping in narrow viewports
- Scorecard CI workflow failing due to incorrect action SHAs
- CodeQL `language:go` analysis missing on PRs (incorrect codeql-action SHA)

### Security
- Patched GO-2026-5024 (`golang.org/x/sys` integer overflow in Windows Unicode handling)
- OpenSSF Scorecard score improved: 5.2 → **7.7**

## [0.1.0] - 2026-06-01

### Added
- CLI tool to scan GitHub repos by language, topic, keyword, and star count
- OpenSSF Scorecard API integration with local CLI fallback
- Web server mode with React frontend embedded via `//go:embed`
- Scan history stored in SQLite (pure-Go driver, no CGo)
- GitHub Trending tab — scores trending repos against Scorecard API
- Single-repo mode (`-single-repo owner/repo`)
- Resizable, sortable results table with sticky header and filter bar
- In-app toast notifications + browser Notification API on scan completion
- Docker / Docker Compose support
- CI/CD pipeline (Build, CodeQL, Scorecard workflows)
- Branch protection on `main`
- OpenSSF Scorecard score: 7.1/10
