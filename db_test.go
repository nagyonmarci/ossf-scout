package main

import (
	"database/sql"
	"path/filepath"
	"testing"
)

func newTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := openDB(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("openDB: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func TestScanCRUD(t *testing.T) {
	db := newTestDB(t)

	cfg := config{
		language:      "go",
		minStars:      1000,
		maxScore:      5.0,
		limit:         50,
		workers:       5,
		checkFilter:   "CI-Tests,SAST",
		cliFallback:   true,
		pushedAfter:   "2026-01-01",
		minMaintained: 3,
		topic:         "kubernetes",
		keyword:       "operator",
		singleRepo:    "",
	}

	id, err := dbInsertScan(db, cfg)
	if err != nil {
		t.Fatalf("dbInsertScan: %v", err)
	}
	if id == 0 {
		t.Fatal("expected non-zero scan id")
	}

	got, err := dbGetScan(db, id)
	if err != nil {
		t.Fatalf("dbGetScan: %v", err)
	}
	if got == nil {
		t.Fatal("dbGetScan returned nil for existing scan")
	}
	if got.Language != "go" || got.MinStars != 1000 || got.Topic != "kubernetes" || got.Keyword != "operator" {
		t.Errorf("scan fields not persisted: %+v", got)
	}
	if got.Status != "running" {
		t.Errorf("new scan status = %q, want running", got.Status)
	}

	// list contains the scan
	scans, err := dbListScans(db)
	if err != nil {
		t.Fatalf("dbListScans: %v", err)
	}
	if len(scans) != 1 || scans[0].ID != id {
		t.Errorf("dbListScans = %+v, want one scan with id %d", scans, id)
	}

	// mark done
	if err := dbUpdateScanDone(db, id, 42, 7); err != nil {
		t.Fatalf("dbUpdateScanDone: %v", err)
	}
	got, _ = dbGetScan(db, id)
	if got.Status != "done" || got.TotalRepos == nil || *got.TotalRepos != 42 || got.ResultCount == nil || *got.ResultCount != 7 {
		t.Errorf("after done: %+v", got)
	}

	// delete
	if err := dbDeleteScan(db, id); err != nil {
		t.Fatalf("dbDeleteScan: %v", err)
	}
	got, err = dbGetScan(db, id)
	if err != nil {
		t.Fatalf("dbGetScan after delete: %v", err)
	}
	if got != nil {
		t.Errorf("scan still present after delete: %+v", got)
	}
}

func TestScanError(t *testing.T) {
	db := newTestDB(t)
	id, err := dbInsertScan(db, config{})
	if err != nil {
		t.Fatalf("dbInsertScan: %v", err)
	}
	if err := dbUpdateScanError(db, id, "boom"); err != nil {
		t.Fatalf("dbUpdateScanError: %v", err)
	}
	got, _ := dbGetScan(db, id)
	if got.Status != "error" || got.ErrorMsg == nil || *got.ErrorMsg != "boom" {
		t.Errorf("after error: %+v", got)
	}
}

func TestGetScanNotFound(t *testing.T) {
	db := newTestDB(t)
	got, err := dbGetScan(db, 99999)
	if err != nil {
		t.Fatalf("dbGetScan: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil for missing scan, got %+v", got)
	}
}

func TestAuditCRUD(t *testing.T) {
	db := newTestDB(t)

	const id = "audit-123"
	if err := dbCreateAudit(db, id, "owner/repo", "opus", "anthropic"); err != nil {
		t.Fatalf("dbCreateAudit: %v", err)
	}

	a, err := dbGetAudit(db, id)
	if err != nil {
		t.Fatalf("dbGetAudit: %v", err)
	}
	if a.Repo != "owner/repo" || a.Status != "pending" || a.Provider != "anthropic" {
		t.Errorf("audit fields not persisted: %+v", a)
	}
	if a.HasContext {
		t.Error("new audit should not have context")
	}

	// save + read context
	if err := dbUpdateAuditContext(db, id, `{"meta":1}`); err != nil {
		t.Fatalf("dbUpdateAuditContext: %v", err)
	}
	ctx, err := dbGetAuditContext(db, id)
	if err != nil {
		t.Fatalf("dbGetAuditContext: %v", err)
	}
	if ctx == nil || *ctx != `{"meta":1}` {
		t.Errorf("context = %v, want saved json", ctx)
	}

	// complete it
	if err := dbUpdateAuditDone(db, id, "# report", 1000, 500); err != nil {
		t.Fatalf("dbUpdateAuditDone: %v", err)
	}
	a, _ = dbGetAudit(db, id)
	if a.Status != "done" || a.Report == nil || *a.Report != "# report" {
		t.Errorf("after done: %+v", a)
	}
	if a.InputTokens == nil || *a.InputTokens != 1000 || a.OutputTokens == nil || *a.OutputTokens != 500 {
		t.Errorf("token counts not persisted: %+v", a)
	}
	if !a.HasContext {
		t.Error("audit should report having context")
	}

	// list
	audits, err := dbListAudits(db)
	if err != nil {
		t.Fatalf("dbListAudits: %v", err)
	}
	if len(audits) != 1 || audits[0].ID != id {
		t.Errorf("dbListAudits = %+v, want one audit %q", audits, id)
	}

	// delete
	if err := dbDeleteAudit(db, id); err != nil {
		t.Fatalf("dbDeleteAudit: %v", err)
	}
	if _, err := dbGetAudit(db, id); err == nil {
		t.Error("expected error fetching deleted audit")
	}
}

func TestAuditErrorWithReport(t *testing.T) {
	db := newTestDB(t)
	const id = "audit-err"
	if err := dbCreateAudit(db, id, "owner/repo", "", ""); err != nil {
		t.Fatalf("dbCreateAudit: %v", err)
	}
	if err := dbUpdateAuditErrorWithReport(db, id, "ai failed", "# fallback snapshot"); err != nil {
		t.Fatalf("dbUpdateAuditErrorWithReport: %v", err)
	}
	a, _ := dbGetAudit(db, id)
	if a.Status != "error" || a.Error == nil || *a.Error != "ai failed" {
		t.Errorf("after error: %+v", a)
	}
	if a.Report == nil || *a.Report != "# fallback snapshot" {
		t.Errorf("fallback report not persisted: %+v", a)
	}
}
