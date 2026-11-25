package scout

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func withScorecardServer(t *testing.T, handler http.HandlerFunc) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	orig := scorecardAPIBase
	scorecardAPIBase = srv.URL
	t.Cleanup(func() { scorecardAPIBase = orig })
}

func TestScorecardGetSuccess(t *testing.T) {
	withScorecardServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/projects/github.com/owner/repo" {
			t.Errorf("path = %q", r.URL.Path)
		}
		if got := r.Header.Get("Accept"); got != "application/json" {
			t.Errorf("Accept = %q", got)
		}
		_, _ = w.Write([]byte(`{"score":7.5,"checks":[{"name":"SAST","score":3}]}`))
	})

	sc, err := scorecardGet(context.Background(), "owner", "repo")
	if err != nil {
		t.Fatalf("scorecardGet: %v", err)
	}
	if sc.Score != 7.5 || len(sc.Checks) != 1 || sc.Checks[0].Name != "SAST" || sc.Checks[0].Score != 3 {
		t.Errorf("unexpected response: %+v", sc)
	}
}

func TestScorecardGetNotFoundReturnsNilNil(t *testing.T) {
	withScorecardServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	sc, err := scorecardGet(context.Background(), "owner", "repo")
	if err != nil {
		t.Fatalf("scorecardGet: %v", err)
	}
	if sc != nil {
		t.Errorf("expected nil response for 404, got %+v", sc)
	}
}

func TestScorecardGetOtherErrorPropagates(t *testing.T) {
	withScorecardServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	if _, err := scorecardGet(context.Background(), "owner", "repo"); err == nil {
		t.Error("expected an error for HTTP 500")
	}
}

func TestScorecardCLIBinaryNotFound(t *testing.T) {
	t.Setenv("PATH", t.TempDir()) // guarantee "scorecard" isn't on PATH

	sc, err := scorecardCLI(context.Background(), "owner", "repo", "")
	if err != nil {
		t.Fatalf("scorecardCLI: %v", err)
	}
	if sc != nil {
		t.Errorf("expected nil response when scorecard binary is missing, got %+v", sc)
	}
}
