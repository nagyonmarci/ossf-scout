package main

import (
	"database/sql"
	"encoding/json"
	"time"

	_ "modernc.org/sqlite"
)

type scanRow struct {
	ID          int64    `json:"id"`
	CreatedAt   string   `json:"created_at"`
	FinishedAt  *string  `json:"finished_at"`
	Status      string   `json:"status"`
	ErrorMsg    *string  `json:"error_msg,omitempty"`
	Language    string   `json:"language"`
	MinStars    int      `json:"min_stars"`
	MaxScore    float64  `json:"max_score"`
	Limit       int      `json:"limit"`
	Workers     int      `json:"workers"`
	CheckFilter string   `json:"check_filter"`
	CliFallback bool     `json:"cli_fallback"`
	PushedAfter   string `json:"pushed_after"`
	MinMaintained int    `json:"min_maintained"`
	Topic         string `json:"topic"`
	Keyword       string `json:"keyword"`
	SingleRepo    string `json:"single_repo"`
	TotalRepos    *int   `json:"total_repos"`
	ResultCount *int     `json:"result_count"`
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
	// Migrate: add columns for existing databases
	_, _ = db.Exec(`ALTER TABLE scan_results ADD COLUMN open_issues INTEGER NOT NULL DEFAULT 0`)
	_, _ = db.Exec(`ALTER TABLE scans ADD COLUMN cli_fallback INTEGER NOT NULL DEFAULT 0`)
	_, _ = db.Exec(`ALTER TABLE scans ADD COLUMN pushed_after TEXT NOT NULL DEFAULT ''`)
	_, _ = db.Exec(`ALTER TABLE scans ADD COLUMN min_maintained INTEGER NOT NULL DEFAULT 0`)
	_, _ = db.Exec(`ALTER TABLE scans ADD COLUMN topic TEXT NOT NULL DEFAULT ''`)
	_, _ = db.Exec(`ALTER TABLE scans ADD COLUMN keyword TEXT NOT NULL DEFAULT ''`)
	_, _ = db.Exec(`ALTER TABLE scans ADD COLUMN single_repo TEXT NOT NULL DEFAULT ''`)
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
	defer stmt.Close()

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
		`SELECT id, created_at, finished_at, status, error_msg,
		        language, min_stars, max_score, limit_, workers, check_filter, cli_fallback, pushed_after, min_maintained, topic, keyword, single_repo,
		        total_repos, result_count
		 FROM scans ORDER BY id DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanRowsFromSQL(rows)
}

func dbGetScan(db *sql.DB, id int64) (*scanRow, error) {
	rows, err := db.Query(
		`SELECT id, created_at, finished_at, status, error_msg,
		        language, min_stars, max_score, limit_, workers, check_filter, cli_fallback, pushed_after, min_maintained, topic, keyword, single_repo,
		        total_repos, result_count
		 FROM scans WHERE id=?`, id,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	scans, err := scanRowsFromSQL(rows)
	if err != nil || len(scans) == 0 {
		return nil, err
	}
	return &scans[0], nil
}

func dbGetResults(db *sql.DB, scanID int64) ([]scanResultRow, error) {
	rows, err := db.Query(
		`SELECT id, scan_id, repo, stars, open_issues, score, language, description,
		        weak_checks, scorecard_url, repo_url
		 FROM scan_results WHERE scan_id=? ORDER BY score ASC`,
		scanID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

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
