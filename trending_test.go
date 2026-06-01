package main

import "testing"

func TestParseTrendingEmpty(t *testing.T) {
	if got := parseTrending(""); len(got) != 0 {
		t.Errorf("expected empty, got %v", got)
	}
}

func TestParseTrendingValid(t *testing.T) {
	html := `<article class="Box-row">
	<h2 class="h3">
		<a href="/golang/go">golang / go</a>
	</h2>
	<span itemprop="programmingLanguage">Go</span>
	<span>1,234 stars today</span>
</article>`
	got := parseTrending(html)
	if len(got) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(got))
	}
	if got[0].FullName != "golang/go" {
		t.Errorf("expected golang/go, got %q", got[0].FullName)
	}
	if got[0].Language != "Go" {
		t.Errorf("expected Go, got %q", got[0].Language)
	}
	if got[0].StarsToday != 1234 {
		t.Errorf("expected 1234 stars, got %d", got[0].StarsToday)
	}
}

func TestParseTrendingMalformed(t *testing.T) {
	cases := []string{
		"<article>no h2 here</article>",
		"<article><h2><a href='/bad'>bad</a></h2></article>",
		"not html at all",
		"<article><!-- unclosed",
	}
	for _, c := range cases {
		got := parseTrending(c)
		_ = got // must not panic
	}
}
