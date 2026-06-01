# Changelog

All notable changes to this project will be documented in this file.
Format based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## [Unreleased]

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
