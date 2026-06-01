package main

import "testing"

func FuzzParseTrending(f *testing.F) {
	f.Add("")
	f.Add("   ")
	f.Add("<article><h2><a href=\"/foo/bar\">foo/bar</a></h2></article>")
	f.Add("<article><h2><a href=\"/a/b\">a/b</a></h2><span itemprop=\"programmingLanguage\">Go</span><span>1,234 stars today</span></article>")
	f.Add("<article><!-- unclosed")
	f.Fuzz(func(t *testing.T, html string) {
		parseTrending(html)
	})
}

func FuzzParseChecks(f *testing.F) {
	f.Add("")
	f.Add("CI-Tests")
	f.Add("CI-Tests,SAST,Branch-Protection")
	f.Add(",,,")
	f.Add("  spaces  ,  trimmed  ")
	f.Fuzz(func(t *testing.T, s string) {
		parseChecks(s)
	})
}
