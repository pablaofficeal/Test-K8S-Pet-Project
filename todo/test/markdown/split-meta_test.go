package markdowntest

import (
	"Cadence-Cloud-Todo/markdown"
	"testing"
)

func TestSplitMeta(t *testing.T) {
	cases := []struct {
		in, title, priority string
		tags int
		deadline string
	}{
		{"plain title", "plain title", "normal", 0, ""},
		{"t | priority:high", "t", "high", 0, ""},
		{"t | priority:HIGH", "t", "high", 0, ""},
		{"t | priority:urgent", "t", "normal", 0, ""},
		{"t| priority:low", "t", "low", 0, ""},
		{"t | tags:a, b ,c | deadline:2026-09-23", "t", "normal", 3, "2026-09-23"},
		{"t | deadline:not-a-date", "t", "normal", 0, ""},
		{"a | b | priority:low", "a | b", "low", 0, ""},
		{"t | priority:low | priority:high", "t", "high", 0, ""},
		{"t | tags:", "t", "normal", 0, ""},
	}
	for _, tc := range cases {
		title, priority, tags, deadline := markdown.SplitMeta(tc.in)
		got := ""
		if deadline != nil {
			got = deadline.Format("2006-01-02")
		}
		if title != tc.title || priority != tc.priority || len(tags) != tc.tags || got != tc.deadline {
			t.Errorf("SplitMeta(%q) = %q %q %v %q", tc.in, title, priority, tags, got)
		}
	}
}