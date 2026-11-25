package scout

import (
	"context"
	"strings"
	"testing"
)

func TestFetchReposInvalidSingleRepo(t *testing.T) {
	// Clearly invalid shapes must error before any network call is made.
	for _, bad := range []string{"bad", "owner/", "/repo"} {
		_, err := New(Config{SingleRepo: bad}).FetchRepos(context.Background())
		if err == nil {
			t.Errorf("FetchRepos(SingleRepo=%q): expected an invalid-format error", bad)
			continue
		}
		if !strings.Contains(err.Error(), "invalid repo format") {
			t.Errorf("FetchRepos(SingleRepo=%q): error = %v", bad, err)
		}
	}
}
