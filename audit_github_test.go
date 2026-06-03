package main

import "testing"

func TestGHListAvailable(t *testing.T) {
	arr := []interface{}{map[string]interface{}{"number": 1.0}}
	cases := []struct {
		status string
		data   interface{}
		want   bool
	}{
		{"ok", arr, true},
		{"ok", nil, true}, // status authoritative
		{"unavailable: API rate limit exceeded", arr, false}, // rate-limit never trusted
		{"unavailable: API rate limit exceeded", nil, false},
		{"", arr, true},  // legacy context with real array
		{"", nil, false}, // legacy, no data
		{"", map[string]interface{}{"message": "x"}, false},
	}
	for _, c := range cases {
		if got := ghListAvailable(c.status, c.data); got != c.want {
			t.Errorf("ghListAvailable(%q, %T)=%v want %v", c.status, c.data, got, c.want)
		}
	}
	if ghStatusLabel("") != "unavailable" {
		t.Error("empty status should label as unavailable")
	}
	if ghStatusLabel("unavailable: rate limit") != "unavailable: rate limit" {
		t.Error("non-empty status should pass through")
	}
}
