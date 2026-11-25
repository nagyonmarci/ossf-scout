package main

import "github.com/nagyonmarci/ossf-scout/internal/scout"

type config struct {
	language    string
	minStars    int
	maxScore    float64
	limit       int
	workers     int
	jsonOut     bool
	token       string
	checkFilter string
	cliFallback   bool
	pushedAfter   string
	minMaintained int
	topic         string
	keyword       string
	singleRepo    string
}

func (c config) toScoutConfig() scout.Config {
	return scout.Config{
		Language:      c.language,
		MinStars:      c.minStars,
		MaxScore:      c.maxScore,
		Limit:         c.limit,
		Workers:       c.workers,
		Token:         c.token,
		CheckFilter:   c.checkFilter,
		CliFallback:   c.cliFallback,
		PushedAfter:   c.pushedAfter,
		MinMaintained: c.minMaintained,
		Topic:         c.topic,
		Keyword:       c.keyword,
		SingleRepo:    c.singleRepo,
	}
}
