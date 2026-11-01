package markdown

import (
	"strings"
	"regexp"
	"time"
)

var metaPattern = regexp.MustCompile(`\s*\|\s*(priority|tags|deadline):`)

func SplitMeta(text string) (string, string, []string, *time.Time) {
	priority := "normal"
	tags := []string{}
	var deadline *time.Time
	found := metaPattern.FindAllStringSubmatchIndex(text, -1)
	if len(found) == 0 {
		return strings.TrimSpace(text), priority, tags, deadline
	}
	title := strings.TrimSpace(text[:found[0][0]])
	for i, m := range found {
		end := len(text)
		if i+1 < len(found) {
			end = found[i+1][0]
		}
		key := text[m[2]:m[3]]
		value := strings.TrimSpace(text[m[1]:end])
		switch key {
		case "priority":
			value = strings.ToLower(value)
			if value == "low" || value == "normal" || value == "high" {
				priority = value
			}
		case "tags":
			tags = []string{}
			for _, tag := range strings.Split(value, ",") {
				if tag = strings.TrimSpace(tag); tag != "" {
					tags = append(tags, tag)
				}
			}
		case "deadline":
			if d, err := time.Parse("2006-01-02", value); err == nil {
				deadline = &d
			}
		}
	}
	return title, priority, tags, deadline
}