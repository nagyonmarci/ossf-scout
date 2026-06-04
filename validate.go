package main

import (
	"fmt"
	"regexp"
)

// reValidRepo matches safe owner/repo strings and captures owner and name separately.
var reValidRepo = regexp.MustCompile(`^([a-zA-Z0-9][a-zA-Z0-9._-]{0,99})/([a-zA-Z0-9][a-zA-Z0-9._-]{0,99})$`)

func validateRepo(repo string) error {
	if !reValidRepo.MatchString(repo) {
		return fmt.Errorf("invalid repository: must be owner/repo using alphanumeric characters, hyphens, dots, or underscores")
	}
	return nil
}

// splitValidRepo validates repo and returns the extracted owner and name components.
// The returned strings come from regex capture groups, not the raw input, so they
// are guaranteed to contain only safe characters.
func splitValidRepo(repo string) (owner, name string, ok bool) {
	m := reValidRepo.FindStringSubmatch(repo)
	if m == nil {
		return "", "", false
	}
	return m[1], m[2], true
}
