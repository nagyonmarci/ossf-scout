package main

type config struct {
	language    string
	minStars    int
	maxScore    float64
	limit       int
	workers     int
	jsonOut     bool
	token       string
	checkFilter string
	cliFallback bool
	pushedAfter string
}
