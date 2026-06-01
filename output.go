package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

func printTable(results []result) {
	fmt.Printf("\n%-50s %6s %8s  %-40s  %s\n",
		"REPOSITORY", "STARS", "SCORE", "WEAK CHECKS", "SCORECARD URL")
	fmt.Println(strings.Repeat("─", 160))
	for _, r := range results {
		score := fmt.Sprintf("%.1f", r.Score)
		if r.Score == -1 {
			score = "N/A"
		}
		fmt.Printf("%-50s %6d %8s  %-40s  %s\n",
			r.Repo.FullName,
			r.Repo.Stars,
			score,
			strings.Join(r.WeakChecks, ", "),
			r.ScorecardURL,
		)
	}
	fmt.Printf("\nFound %d repos matching criteria.\n", len(results))
}

type jsonResult struct {
	Repo         string   `json:"repo"`
	Stars        int      `json:"stars"`
	Score        float64  `json:"score"`
	Language     string   `json:"language"`
	Description  string   `json:"description"`
	WeakChecks   []string `json:"weak_checks"`
	ScorecardURL string   `json:"scorecard_url"`
	RepoURL      string   `json:"repo_url"`
}

func printJSON(results []result) {
	var out []jsonResult
	for _, r := range results {
		out = append(out, jsonResult{
			Repo:         r.Repo.FullName,
			Stars:        r.Repo.Stars,
			Score:        r.Score,
			Language:     r.Repo.Language,
			Description:  r.Repo.Description,
			WeakChecks:   r.WeakChecks,
			ScorecardURL: r.ScorecardURL,
			RepoURL:      r.Repo.HTMLURL,
		})
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	_ = enc.Encode(out)
}
