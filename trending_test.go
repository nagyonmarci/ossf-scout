package main

import (
	"fmt"
	"strings"
	"testing"
)

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

func TestParseTrendingMultiple(t *testing.T) {
	html := `<article><h2><a href="/golang/go">go</a></h2><span itemprop="programmingLanguage">Go</span><span>1,234 stars today</span></article>
<article><h2><a href="/rust-lang/rust">rust</a></h2><span itemprop="programmingLanguage">Rust</span><span>900 stars today</span></article>
<article><h2><a href="/torvalds/linux">linux</a></h2></article>`
	got := parseTrending(html)
	if len(got) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(got))
	}
	if got[1].FullName != "rust-lang/rust" || got[1].Language != "Rust" || got[1].StarsToday != 900 {
		t.Errorf("entry[1] = %+v", got[1])
	}
}

func TestParseTrendingMissingOptionalFields(t *testing.T) {
	// Repo link present but no language and no star line → defaults, still parsed.
	html := `<article><h2><a href="/owner/repo">x</a></h2></article>`
	got := parseTrending(html)
	if len(got) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(got))
	}
	if got[0].FullName != "owner/repo" || got[0].Language != "" || got[0].StarsToday != 0 {
		t.Errorf("entry = %+v, want owner/repo with empty lang and 0 stars", got[0])
	}
}

func TestParseTrendingTimeWindows(t *testing.T) {
	cases := map[string]int{
		"<span>1,234 stars today</span>":     1234,
		"<span>56 stars this week</span>":    56,
		"<span>78 stars this month</span>":   78,
		"<span>9,012 stars this year</span>": 9012,
	}
	for starLine, want := range cases {
		html := `<article><h2><a href="/o/r">r</a></h2>` + starLine + `</article>`
		got := parseTrending(html)
		if len(got) != 1 {
			t.Fatalf("%q: expected 1 entry, got %d", starLine, len(got))
		}
		if got[0].StarsToday != want {
			t.Errorf("%q: stars = %d, want %d", starLine, got[0].StarsToday, want)
		}
	}
}

func TestParseTrendingCapsArticleCount(t *testing.T) {
	var b strings.Builder
	for i := 0; i < maxTrendingArticles+50; i++ {
		fmt.Fprintf(&b, `<article><h2><a href="/owner/repo%d">x</a></h2></article>`, i)
	}
	got := parseTrending(b.String())
	if len(got) != maxTrendingArticles {
		t.Errorf("expected results capped at %d, got %d", maxTrendingArticles, len(got))
	}
}
