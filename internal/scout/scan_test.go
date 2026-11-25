package scout

import (
	"context"
	"encoding/json"
	"flag"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

var update = flag.Bool("update", false, "update golden files")

func TestScanGolden(t *testing.T) {
	ghSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("page") != "1" {
			_, _ = w.Write([]byte(`{"items":[]}`))
			return
		}
		_, _ = w.Write([]byte(`{"items":[
			{"full_name":"acme/widget","description":"widgets","stargazers_count":600,"html_url":"https://github.com/acme/widget","language":"Go","topics":["go"]},
			{"full_name":"acme/gadget","description":"gadgets","stargazers_count":900,"html_url":"https://github.com/acme/gadget","language":"Go","topics":["go"]}
		]}`))
	}))
	defer ghSrv.Close()
	origGH := githubAPIBase
	githubAPIBase = ghSrv.URL
	defer func() { githubAPIBase = origGH }()

	scSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/projects/github.com/acme/widget":
			_, _ = w.Write([]byte(`{"score":3.2,"checks":[{"name":"SAST","score":0},{"name":"Maintained","score":8}]}`))
		case "/projects/github.com/acme/gadget":
			_, _ = w.Write([]byte(`{"score":8.9,"checks":[{"name":"SAST","score":9},{"name":"Maintained","score":9}]}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer scSrv.Close()
	origSC := scorecardAPIBase
	scorecardAPIBase = scSrv.URL
	defer func() { scorecardAPIBase = origSC }()

	origDelay := searchPageDelay
	searchPageDelay = 0
	defer func() { searchPageDelay = origDelay }()

	sc := New(Config{MinStars: 500, MaxScore: 5.0, Limit: 10, Workers: 2})
	results, err := sc.Scan(context.Background())
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}

	got, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	got = append(got, '\n')

	goldenPath := filepath.Join("testdata", "scan_golden.json")
	if *update {
		if err := os.WriteFile(goldenPath, got, 0o644); err != nil {
			t.Fatalf("write golden: %v", err)
		}
	}
	want, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("read golden (run with -update to generate it): %v", err)
	}
	if string(got) != string(want) {
		t.Errorf("Scan result mismatch (run with -update to regenerate):\ngot:  %s\nwant: %s", got, want)
	}
}
