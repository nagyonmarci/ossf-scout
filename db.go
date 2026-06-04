package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

type scanRow struct {
	ID            int64   `json:"id"`
	CreatedAt     string  `json:"created_at"`
	FinishedAt    *string `json:"finished_at"`
	Status        string  `json:"status"`
	ErrorMsg      *string `json:"error_msg,omitempty"`
	Language      string  `json:"language"`
	MinStars      int     `json:"min_stars"`
	MaxScore      float64 `json:"max_score"`
	Limit         int     `json:"limit"`
	Workers       int     `json:"workers"`
	CheckFilter   string  `json:"check_filter"`
	CliFallback   bool    `json:"cli_fallback"`
	PushedAfter   string  `json:"pushed_after"`
	MinMaintained int     `json:"min_maintained"`
	Topic         string  `json:"topic"`
	Keyword       string  `json:"keyword"`
	SingleRepo    string  `json:"single_repo"`
	TotalRepos    *int    `json:"total_repos"`
	ResultCount   *int    `json:"result_count"`
}

type auditRow struct {
	ID           string     `json:"id"`
	Repo         string     `json:"repo"`
	Status       string     `json:"status"`
	Model        string     `json:"model"`
	Provider     string     `json:"provider"`
	CreatedAt    time.Time  `json:"created_at"`
	CompletedAt  *time.Time `json:"completed_at"`
	Report       *string    `json:"report"`
	Error        *string    `json:"error"`
	InputTokens  *int       `json:"input_tokens"`
	OutputTokens *int       `json:"output_tokens"`
	HasContext   bool       `json:"has_context"`
}

type scanResultRow struct {
	ID           int64    `json:"id"`
	ScanID       int64    `json:"scan_id"`
	Repo         string   `json:"repo"`
	Stars        int      `json:"stars"`
	OpenIssues   int      `json:"open_issues"`
	Score        float64  `json:"score"`
	Language     string   `json:"language"`
	Description  string   `json:"description"`
	WeakChecks   []string `json:"weak_checks"`
	ScorecardURL string   `json:"scorecard_url"`
	RepoURL      string   `json:"repo_url"`
}

const (
	scanColumns = `id, created_at, finished_at, status, error_msg,
	        language, min_stars, max_score, limit_, workers, check_filter, cli_fallback, pushed_after, min_maintained, topic, keyword, single_repo,
	        total_repos, result_count`
	resultColumns = `id, scan_id, repo, stars, open_issues, score, language, description,
	        weak_checks, scorecard_url, repo_url`
)

const schema = `
CREATE TABLE IF NOT EXISTS scans (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    created_at   DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
    finished_at  DATETIME,
    status       TEXT NOT NULL DEFAULT 'running',
    error_msg    TEXT,
    language     TEXT NOT NULL DEFAULT '',
    min_stars    INTEGER NOT NULL DEFAULT 500,
    max_score    REAL NOT NULL DEFAULT 5.0,
    limit_       INTEGER NOT NULL DEFAULT 100,
    workers      INTEGER NOT NULL DEFAULT 5,
    check_filter TEXT NOT NULL DEFAULT '',
    cli_fallback INTEGER NOT NULL DEFAULT 0,
    pushed_after   TEXT NOT NULL DEFAULT '',
    min_maintained INTEGER NOT NULL DEFAULT 0,
    topic          TEXT NOT NULL DEFAULT '',
    keyword        TEXT NOT NULL DEFAULT '',
    single_repo    TEXT NOT NULL DEFAULT '',
    total_repos    INTEGER,
    result_count INTEGER
);

CREATE TABLE IF NOT EXISTS scan_results (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    scan_id       INTEGER NOT NULL REFERENCES scans(id) ON DELETE CASCADE,
    repo          TEXT NOT NULL,
    stars         INTEGER NOT NULL,
    open_issues   INTEGER NOT NULL DEFAULT 0,
    score         REAL NOT NULL,
    language      TEXT NOT NULL DEFAULT '',
    description   TEXT NOT NULL DEFAULT '',
    weak_checks   TEXT NOT NULL DEFAULT '[]',
    scorecard_url TEXT NOT NULL DEFAULT '',
    repo_url      TEXT NOT NULL DEFAULT ''
);

CREATE INDEX IF NOT EXISTS idx_scan_results_scan_id ON scan_results(scan_id);
`

func openDB(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	if _, err := db.Exec("PRAGMA foreign_keys = ON;PRAGMA journal_mode = WAL;"); err != nil {
		return nil, err
	}
	if _, err := db.Exec(schema); err != nil {
		return nil, err
	}
	const auditSchema = `
CREATE TABLE IF NOT EXISTS audits (
    id            TEXT    NOT NULL PRIMARY KEY,
    repo          TEXT    NOT NULL,
    status        TEXT    NOT NULL DEFAULT 'pending',
    model         TEXT    NOT NULL DEFAULT '',
    provider      TEXT    NOT NULL DEFAULT '',
    created_at    DATETIME NOT NULL,
    completed_at  DATETIME,
    report        TEXT,
    error         TEXT,
    input_tokens  INTEGER,
    output_tokens INTEGER
);`
	if _, err := db.Exec(auditSchema); err != nil {
		return nil, fmt.Errorf("create audits table: %w", err)
	}
	// Migrate: add columns for existing databases
	_, _ = db.Exec(`ALTER TABLE scan_results ADD COLUMN open_issues INTEGER NOT NULL DEFAULT 0`)
	_, _ = db.Exec(`ALTER TABLE scans ADD COLUMN cli_fallback INTEGER NOT NULL DEFAULT 0`)
	_, _ = db.Exec(`ALTER TABLE scans ADD COLUMN pushed_after TEXT NOT NULL DEFAULT ''`)
	_, _ = db.Exec(`ALTER TABLE scans ADD COLUMN min_maintained INTEGER NOT NULL DEFAULT 0`)
	_, _ = db.Exec(`ALTER TABLE scans ADD COLUMN topic TEXT NOT NULL DEFAULT ''`)
	_, _ = db.Exec(`ALTER TABLE scans ADD COLUMN keyword TEXT NOT NULL DEFAULT ''`)
	_, _ = db.Exec(`ALTER TABLE scans ADD COLUMN single_repo TEXT NOT NULL DEFAULT ''`)
	_, _ = db.Exec(`ALTER TABLE audits ADD COLUMN model TEXT NOT NULL DEFAULT ''`)
	_, _ = db.Exec(`ALTER TABLE audits ADD COLUMN provider TEXT NOT NULL DEFAULT ''`)
	_, _ = db.Exec(`ALTER TABLE audits ADD COLUMN context_json TEXT`)
	_, _ = db.Exec(`ALTER TABLE audits ADD COLUMN head_sha TEXT NOT NULL DEFAULT ''`)

	const scheduleSchema = `
CREATE TABLE IF NOT EXISTS audit_schedules (
    id                TEXT PRIMARY KEY,
    repo              TEXT NOT NULL,
    interval_h        INTEGER NOT NULL DEFAULT 168,
    provider          TEXT NOT NULL DEFAULT '',
    model             TEXT NOT NULL DEFAULT '',
    enabled           INTEGER NOT NULL DEFAULT 1,
    time_window_start INTEGER NOT NULL DEFAULT -1,
    time_window_end   INTEGER NOT NULL DEFAULT -1,
    cli_fallback      INTEGER NOT NULL DEFAULT 1,
    auto_detected     INTEGER NOT NULL DEFAULT 0,
    detect_reason     TEXT NOT NULL DEFAULT '',
    last_run_at       DATETIME,
    next_run_at       DATETIME NOT NULL,
    created_at        DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now'))
);
CREATE INDEX IF NOT EXISTS idx_schedules_next_run ON audit_schedules(next_run_at);
CREATE INDEX IF NOT EXISTS idx_schedules_repo ON audit_schedules(repo);`
	if _, err := db.Exec(scheduleSchema); err != nil {
		return nil, fmt.Errorf("create audit_schedules table: %w", err)
	}

	const issuePRsSchema = `
CREATE TABLE IF NOT EXISTS repo_issues_prs (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    repo       TEXT NOT NULL,
    fetched_at DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
    data_json  TEXT NOT NULL,
    summary_md TEXT
);
CREATE INDEX IF NOT EXISTS idx_repo_issues_prs_repo ON repo_issues_prs(repo);`
	if _, err := db.Exec(issuePRsSchema); err != nil {
		return nil, fmt.Errorf("create repo_issues_prs table: %w", err)
	}

	// Mark any scans that were running when the server last died
	_, _ = db.Exec(`UPDATE scans SET status='error', error_msg='server restarted' WHERE status='running'`)
	return db, nil
}

func dbInsertScan(db *sql.DB, cfg config) (int64, error) {
	res, err := db.Exec(
		`INSERT INTO scans (language, min_stars, max_score, limit_, workers, check_filter, cli_fallback, pushed_after, min_maintained, topic, keyword, single_repo)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		cfg.language, cfg.minStars, cfg.maxScore, cfg.limit, cfg.workers, cfg.checkFilter, cfg.cliFallback, cfg.pushedAfter, cfg.minMaintained, cfg.topic, cfg.keyword, cfg.singleRepo,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func dbUpdateScanDone(db *sql.DB, id int64, totalRepos, resultCount int) error {
	_, err := db.Exec(
		`UPDATE scans SET status='done', finished_at=?, total_repos=?, result_count=? WHERE id=?`,
		time.Now().UTC().Format(time.RFC3339Nano), totalRepos, resultCount, id,
	)
	return err
}

func dbUpdateScanError(db *sql.DB, id int64, msg string) error {
	_, err := db.Exec(
		`UPDATE scans SET status='error', finished_at=?, error_msg=? WHERE id=?`,
		time.Now().UTC().Format(time.RFC3339Nano), msg, id,
	)
	return err
}

func dbInsertResults(db *sql.DB, scanID int64, results []result) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare(
		`INSERT INTO scan_results (scan_id, repo, stars, open_issues, score, language, description, weak_checks, scorecard_url, repo_url)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
	)
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	defer stmt.Close() //nolint:errcheck

	for _, r := range results {
		wc, _ := json.Marshal(r.WeakChecks)
		_, err := stmt.Exec(
			scanID, r.Repo.FullName, r.Repo.Stars, r.Repo.OpenIssuesCount, r.Score,
			r.Repo.Language, r.Repo.Description, string(wc),
			r.ScorecardURL, r.Repo.HTMLURL,
		)
		if err != nil {
			_ = tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

func dbListScans(db *sql.DB) ([]scanRow, error) {
	rows, err := db.Query(
		`SELECT ` + scanColumns + ` FROM scans ORDER BY id DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close() //nolint:errcheck
	return scanRowsFromSQL(rows)
}

func dbGetScan(db *sql.DB, id int64) (*scanRow, error) {
	rows, err := db.Query(
		`SELECT `+scanColumns+` FROM scans WHERE id=?`, id,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close() //nolint:errcheck
	scans, err := scanRowsFromSQL(rows)
	if err != nil || len(scans) == 0 {
		return nil, err
	}
	return &scans[0], nil
}

func dbGetResults(db *sql.DB, scanID int64) ([]scanResultRow, error) {
	rows, err := db.Query(
		`SELECT `+resultColumns+` FROM scan_results WHERE scan_id=? ORDER BY score ASC`,
		scanID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close() //nolint:errcheck

	var out []scanResultRow
	for rows.Next() {
		var r scanResultRow
		var wcJSON string
		err := rows.Scan(&r.ID, &r.ScanID, &r.Repo, &r.Stars, &r.OpenIssues, &r.Score,
			&r.Language, &r.Description, &wcJSON, &r.ScorecardURL, &r.RepoURL)
		if err != nil {
			return nil, err
		}
		_ = json.Unmarshal([]byte(wcJSON), &r.WeakChecks)
		if r.WeakChecks == nil {
			r.WeakChecks = []string{}
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func dbDeleteScan(db *sql.DB, id int64) error {
	_, err := db.Exec(`DELETE FROM scans WHERE id=?`, id)
	return err
}

func dbCreateAudit(db *sql.DB, id, repo, model, provider string) error {
	_, err := db.Exec(
		`INSERT INTO audits (id, repo, status, model, provider, created_at) VALUES (?, ?, 'pending', ?, ?, ?)`,
		id, repo, model, provider, time.Now().UTC(),
	)
	return err
}

func dbUpdateAuditRunning(db *sql.DB, id string) error {
	_, err := db.Exec(`UPDATE audits SET status='running' WHERE id=?`, id)
	return err
}

func dbUpdateAuditContext(db *sql.DB, id, contextJSON string) error {
	_, err := db.Exec(`UPDATE audits SET context_json=? WHERE id=?`, contextJSON, id)
	return err
}

func dbGetAuditContext(db *sql.DB, id string) (*string, error) {
	var ctx *string
	err := db.QueryRow(`SELECT context_json FROM audits WHERE id=?`, id).Scan(&ctx)
	if err != nil {
		return nil, err
	}
	return ctx, nil
}

func dbRestartAudit(db *sql.DB, id, model, provider string) error {
	_, err := db.Exec(
		`UPDATE audits SET status='running', model=?, provider=?, error=NULL WHERE id=?`,
		model, provider, id,
	)
	return err
}

func dbUpdateAuditDone(db *sql.DB, id, report string, inputTokens, outputTokens int) error {
	_, err := db.Exec(
		`UPDATE audits SET status='done', completed_at=?, report=?, input_tokens=?, output_tokens=? WHERE id=?`,
		time.Now().UTC(), report, inputTokens, outputTokens, id,
	)
	return err
}

func dbUpdateAuditError(db *sql.DB, id, errMsg string) error {
	_, err := db.Exec(
		`UPDATE audits SET status='error', completed_at=?, error=? WHERE id=?`,
		time.Now().UTC(), errMsg, id,
	)
	return err
}

func dbUpdateAuditErrorWithReport(db *sql.DB, id, errMsg, report string) error {
	_, err := db.Exec(
		`UPDATE audits SET status='error', completed_at=?, error=?, report=? WHERE id=?`,
		time.Now().UTC(), errMsg, report, id,
	)
	return err
}

func dbListAudits(db *sql.DB) ([]auditRow, error) {
	rows, err := db.Query(
		`SELECT id, repo, status, model, provider, created_at, completed_at, error, input_tokens, output_tokens,
		        CASE WHEN context_json IS NULL THEN 0 ELSE 1 END
		 FROM audits ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close() //nolint:errcheck
	var out []auditRow
	for rows.Next() {
		var a auditRow
		var hasCtx int
		if err := rows.Scan(&a.ID, &a.Repo, &a.Status, &a.Model, &a.Provider, &a.CreatedAt,
			&a.CompletedAt, &a.Error, &a.InputTokens, &a.OutputTokens, &hasCtx); err != nil {
			return nil, err
		}
		a.HasContext = hasCtx != 0
		out = append(out, a)
	}
	return out, rows.Err()
}

func dbGetAudit(db *sql.DB, id string) (*auditRow, error) {
	var a auditRow
	var hasCtx int
	err := db.QueryRow(
		`SELECT id, repo, status, model, provider, created_at, completed_at, report, error, input_tokens, output_tokens,
		        CASE WHEN context_json IS NULL THEN 0 ELSE 1 END
		 FROM audits WHERE id=?`, id,
	).Scan(&a.ID, &a.Repo, &a.Status, &a.Model, &a.Provider, &a.CreatedAt,
		&a.CompletedAt, &a.Report, &a.Error, &a.InputTokens, &a.OutputTokens, &hasCtx)
	if err != nil {
		return nil, err
	}
	a.HasContext = hasCtx != 0
	return &a, nil
}

func dbDeleteAudit(db *sql.DB, id string) error {
	_, err := db.Exec(`DELETE FROM audits WHERE id=?`, id)
	return err
}

// ── SHA-cache ────────────────────────────────────────────────────────────────

func dbUpdateAuditHeadSHA(db *sql.DB, id, sha string) error {
	_, err := db.Exec(`UPDATE audits SET head_sha=? WHERE id=?`, sha, id)
	return err
}

// dbFindRecentContext returns context_json for the most recent completed audit
// of the same repo+SHA within maxAge, or nil if none found.
func dbFindRecentContext(db *sql.DB, repo, sha string, maxAge time.Duration) (*string, error) {
	if sha == "" {
		return nil, nil
	}
	cutoff := time.Now().UTC().Add(-maxAge)
	var ctx string
	err := db.QueryRow(
		`SELECT context_json FROM audits
		 WHERE repo=? AND head_sha=? AND context_json IS NOT NULL
		   AND created_at >= ?
		 ORDER BY created_at DESC LIMIT 1`,
		repo, sha, cutoff.Format(time.RFC3339),
	).Scan(&ctx)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &ctx, nil
}

// ── Cost stats ───────────────────────────────────────────────────────────────

type costStats struct {
	TotalUSD      float64            `json:"total_usd"`
	ByModel       map[string]float64 `json:"by_model"`
	InputTokens   int                `json:"total_input_tokens"`
	OutputTokens  int                `json:"total_output_tokens"`
	AuditCount    int                `json:"audit_count"`
}

func dbGetCostStats(db *sql.DB, days int) (costStats, error) {
	cutoff := time.Now().UTC().AddDate(0, 0, -days).Format(time.RFC3339)
	rows, err := db.Query(
		`SELECT model, COALESCE(input_tokens,0), COALESCE(output_tokens,0)
		 FROM audits
		 WHERE status='done' AND created_at >= ?`, cutoff,
	)
	if err != nil {
		return costStats{}, err
	}
	defer rows.Close() //nolint:errcheck

	s := costStats{ByModel: make(map[string]float64)}
	for rows.Next() {
		var model string
		var in, out int
		if err := rows.Scan(&model, &in, &out); err != nil {
			return costStats{}, err
		}
		s.InputTokens += in
		s.OutputTokens += out
		s.AuditCount++
		if prices, ok := modelPrices[model]; ok {
			cost := float64(in)/1_000_000*prices[0] + float64(out)/1_000_000*prices[1]
			s.ByModel[model] += cost
			s.TotalUSD += cost
		}
	}
	return s, rows.Err()
}

// ── Schedules ────────────────────────────────────────────────────────────────

type scheduleRow struct {
	ID              string     `json:"id"`
	Repo            string     `json:"repo"`
	IntervalH       int        `json:"interval_h"`
	Provider        string     `json:"provider"`
	Model           string     `json:"model"`
	Enabled         bool       `json:"enabled"`
	TimeWindowStart int        `json:"time_window_start"`
	TimeWindowEnd   int        `json:"time_window_end"`
	CliFallback     bool       `json:"cli_fallback"`
	AutoDetected    bool       `json:"auto_detected"`
	DetectReason    string     `json:"detect_reason,omitempty"`
	LastRunAt       *time.Time `json:"last_run_at"`
	NextRunAt       time.Time  `json:"next_run_at"`
	CreatedAt       time.Time  `json:"created_at"`
}

const scheduleColumns = `id, repo, interval_h, provider, model, enabled,
	time_window_start, time_window_end, cli_fallback, auto_detected, detect_reason,
	last_run_at, next_run_at, created_at`

func scheduleRowsFromSQL(rows *sql.Rows) ([]scheduleRow, error) {
	var out []scheduleRow
	for rows.Next() {
		var s scheduleRow
		var enabled, cliFallback, autoDetected int
		if err := rows.Scan(
			&s.ID, &s.Repo, &s.IntervalH, &s.Provider, &s.Model, &enabled,
			&s.TimeWindowStart, &s.TimeWindowEnd, &cliFallback, &autoDetected, &s.DetectReason,
			&s.LastRunAt, &s.NextRunAt, &s.CreatedAt,
		); err != nil {
			return nil, err
		}
		s.Enabled = enabled != 0
		s.CliFallback = cliFallback != 0
		s.AutoDetected = autoDetected != 0
		out = append(out, s)
	}
	return out, rows.Err()
}

func dbListSchedules(db *sql.DB) ([]scheduleRow, error) {
	rows, err := db.Query(`SELECT ` + scheduleColumns + ` FROM audit_schedules ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close() //nolint:errcheck
	return scheduleRowsFromSQL(rows)
}

func dbListDueSchedules(db *sql.DB, now time.Time) ([]scheduleRow, error) {
	rows, err := db.Query(
		`SELECT `+scheduleColumns+` FROM audit_schedules
		 WHERE enabled=1 AND next_run_at <= ?`,
		now.UTC().Format(time.RFC3339),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close() //nolint:errcheck
	return scheduleRowsFromSQL(rows)
}

func dbCreateSchedule(db *sql.DB, id, repo string, intervalH int, provider, model string, timeWindowStart, timeWindowEnd int, cliFallback, autoDetected bool, detectReason string) error {
	nextRun := time.Now().UTC().Add(time.Duration(intervalH) * time.Hour)
	_, err := db.Exec(
		`INSERT OR IGNORE INTO audit_schedules
		 (id, repo, interval_h, provider, model, time_window_start, time_window_end, cli_fallback, auto_detected, detect_reason, next_run_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		id, repo, intervalH, provider, model, timeWindowStart, timeWindowEnd,
		boolToInt(cliFallback), boolToInt(autoDetected), detectReason,
		nextRun.Format(time.RFC3339),
	)
	return err
}

func dbUpdateSchedule(db *sql.DB, id string, intervalH int, provider, model string, enabled bool, timeWindowStart, timeWindowEnd int) error {
	_, err := db.Exec(
		`UPDATE audit_schedules SET interval_h=?, provider=?, model=?, enabled=?, time_window_start=?, time_window_end=? WHERE id=?`,
		intervalH, provider, model, boolToInt(enabled), timeWindowStart, timeWindowEnd, id,
	)
	return err
}

func dbUpdateScheduleLastRun(db *sql.DB, id string, now time.Time, nextRun time.Time) error {
	_, err := db.Exec(
		`UPDATE audit_schedules SET last_run_at=?, next_run_at=? WHERE id=?`,
		now.UTC().Format(time.RFC3339), nextRun.UTC().Format(time.RFC3339), id,
	)
	return err
}

func dbDeleteSchedule(db *sql.DB, id string) error {
	_, err := db.Exec(`DELETE FROM audit_schedules WHERE id=?`, id)
	return err
}

func dbTriggerScheduleNow(db *sql.DB, id string) error {
	_, err := db.Exec(`UPDATE audit_schedules SET next_run_at=? WHERE id=?`,
		time.Now().UTC().Add(-time.Second).Format(time.RFC3339), id)
	return err
}

// dbAutoDetectSchedules creates schedule suggestions for repos that warrant monitoring.
// Rule 1: scan_results with score < 5.0 AND stars > 500 → enabled=1, weekly
// Rule 2: repos with 3+ audits in 90 days → enabled=0 suggestion
func dbAutoDetectSchedules(db *sql.DB) error {
	// Rule 1: high-star, low-score repos from scan_results
	rows, err := db.Query(
		`SELECT DISTINCT repo FROM scan_results
		 WHERE score < 5.0 AND stars > 500
		   AND repo NOT IN (SELECT repo FROM audit_schedules)
		 LIMIT 50`,
	)
	if err != nil {
		return err
	}
	defer rows.Close() //nolint:errcheck
	var repos []string
	for rows.Next() {
		var r string
		if err := rows.Scan(&r); err != nil {
			return err
		}
		repos = append(repos, r)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	for _, repo := range repos {
		id := repo // deterministic ID to avoid duplicates across restarts
		reason := "high-star repo with low Scorecard score (auto-detected)"
		if err := dbCreateSchedule(db, "auto-"+hashStr(id), repo,
			defaultScheduleIntervalH, "", "", -1, -1, true, true, reason); err != nil {
			continue
		}
	}

	// Rule 2: repos audited 3+ times in 90 days → suggest (enabled=0)
	cutoff := time.Now().UTC().AddDate(0, 0, -90).Format(time.RFC3339)
	rows2, err := db.Query(
		`SELECT repo, COUNT(*) as cnt FROM audits
		 WHERE created_at >= ?
		   AND repo NOT IN (SELECT repo FROM audit_schedules)
		 GROUP BY repo HAVING cnt >= 3 LIMIT 20`, cutoff,
	)
	if err != nil {
		return err
	}
	defer rows2.Close() //nolint:errcheck
	for rows2.Next() {
		var repo string
		var cnt int
		if err := rows2.Scan(&repo, &cnt); err != nil {
			return err
		}
		reason := fmt.Sprintf("audited %d times in the last 90 days (suggested)", cnt)
		sid := "suggest-" + hashStr(repo)
		_ = dbCreateSchedule(db, sid, repo, defaultScheduleIntervalH, "", "", -1, -1, false, true, reason)
		// disable the suggestion
		_, _ = db.Exec(`UPDATE audit_schedules SET enabled=0 WHERE id=?`, sid)
	}
	return rows2.Err()
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// hashStr returns a short deterministic hex string for use as ID suffix.
func hashStr(s string) string {
	h := uint32(2166136261)
	for i := 0; i < len(s); i++ {
		h ^= uint32(s[i])
		h *= 16777619
	}
	return fmt.Sprintf("%08x", h)
}

func scanRowsFromSQL(rows *sql.Rows) ([]scanRow, error) {
	var out []scanRow
	for rows.Next() {
		var s scanRow
		err := rows.Scan(
			&s.ID, &s.CreatedAt, &s.FinishedAt, &s.Status, &s.ErrorMsg,
			&s.Language, &s.MinStars, &s.MaxScore, &s.Limit, &s.Workers, &s.CheckFilter, &s.CliFallback, &s.PushedAfter, &s.MinMaintained, &s.Topic, &s.Keyword, &s.SingleRepo,
			&s.TotalRepos, &s.ResultCount,
		)
		if err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}
