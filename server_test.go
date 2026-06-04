package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

// testScanMux wires up the read-only scan endpoints against a real (temp) DB.
// It deliberately omits POST /api/scans because that handler spawns a
// background scan goroutine; the read/delete paths are fully hermetic.
func testScanMux(db *sql.DB) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/scans", handleListScans(db))
	mux.HandleFunc("GET /api/scans/{id}", handleGetScan(db))
	mux.HandleFunc("GET /api/scans/{id}/results", handleGetResults(db))
	mux.HandleFunc("DELETE /api/scans/{id}", handleDeleteScan(db))
	return mux
}

func TestHandleListScansEmpty(t *testing.T) {
	db := newTestDB(t)
	srv := httptest.NewServer(testScanMux(db))
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/scans")
	if err != nil {
		t.Fatalf("GET /api/scans: %v", err)
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	var scans []scanRow
	if err := json.NewDecoder(resp.Body).Decode(&scans); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if scans == nil {
		t.Error("empty list should serialize as [] not null")
	}
	if len(scans) != 0 {
		t.Errorf("expected 0 scans, got %d", len(scans))
	}
}

func TestHandleGetScan(t *testing.T) {
	db := newTestDB(t)
	id, err := dbInsertScan(db, config{language: "go", minStars: 500})
	if err != nil {
		t.Fatalf("seed scan: %v", err)
	}
	srv := httptest.NewServer(testScanMux(db))
	defer srv.Close()

	// existing scan → 200
	resp, err := http.Get(srv.URL + "/api/scans/" + itoa(id))
	if err != nil {
		t.Fatalf("GET scan: %v", err)
	}
	defer resp.Body.Close() //nolint:errcheck
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	var s scanRow
	if err := json.NewDecoder(resp.Body).Decode(&s); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if s.ID != id || s.Language != "go" {
		t.Errorf("unexpected scan: %+v", s)
	}
}

func TestHandleGetScanNotFound(t *testing.T) {
	db := newTestDB(t)
	srv := httptest.NewServer(testScanMux(db))
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/scans/99999")
	if err != nil {
		t.Fatalf("GET missing scan: %v", err)
	}
	defer resp.Body.Close() //nolint:errcheck
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("status = %d, want 404", resp.StatusCode)
	}
}

func TestHandleGetScanBadID(t *testing.T) {
	db := newTestDB(t)
	srv := httptest.NewServer(testScanMux(db))
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/scans/not-a-number")
	if err != nil {
		t.Fatalf("GET bad id: %v", err)
	}
	defer resp.Body.Close() //nolint:errcheck
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", resp.StatusCode)
	}
}

func TestHandleDeleteScan(t *testing.T) {
	db := newTestDB(t)
	id, err := dbInsertScan(db, config{})
	if err != nil {
		t.Fatalf("seed scan: %v", err)
	}
	srv := httptest.NewServer(testScanMux(db))
	defer srv.Close()

	req, _ := http.NewRequest(http.MethodDelete, srv.URL+"/api/scans/"+itoa(id), nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("DELETE scan: %v", err)
	}
	defer resp.Body.Close() //nolint:errcheck
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", resp.StatusCode)
	}

	got, _ := dbGetScan(db, id)
	if got != nil {
		t.Errorf("scan still present after DELETE: %+v", got)
	}
}

func TestHandleGetResultsEmpty(t *testing.T) {
	db := newTestDB(t)
	id, err := dbInsertScan(db, config{})
	if err != nil {
		t.Fatalf("seed scan: %v", err)
	}
	srv := httptest.NewServer(testScanMux(db))
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/scans/" + itoa(id) + "/results")
	if err != nil {
		t.Fatalf("GET results: %v", err)
	}
	defer resp.Body.Close() //nolint:errcheck
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	var results []scanResultRow
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if results == nil {
		t.Error("empty results should serialize as [] not null")
	}
}

func itoa(n int64) string {
	return strconv.FormatInt(n, 10)
}
