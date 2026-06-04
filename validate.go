package main

import (
	"fmt"
	"net/url"
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

// validateOllamaURL checks that the Ollama base URL is a well-formed http/https URL.
func validateOllamaURL(rawURL string) error {
	u, err := url.Parse(rawURL)
	if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") {
		return fmt.Errorf("invalid Ollama URL: must be an http:// or https:// URL")
	}
	return nil
}
