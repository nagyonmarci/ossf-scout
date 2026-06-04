package main

import (
	"fmt"
	"regexp"
)

// reValidRepo matches safe owner/repo strings (alphanumeric, hyphen, dot, underscore only).
// This prevents shell-metacharacter injection into git CLI arguments.
var reValidRepo = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._-]{0,99}/[a-zA-Z0-9][a-zA-Z0-9._-]{0,99}$`)

func validateRepo(repo string) error {
	if !reValidRepo.MatchString(repo) {
		return fmt.Errorf("invalid repository: must be owner/repo using alphanumeric characters, hyphens, dots, or underscores")
	}
	return nil
}
