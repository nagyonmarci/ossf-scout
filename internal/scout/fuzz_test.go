package scout

import "testing"

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
