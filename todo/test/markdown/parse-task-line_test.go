package markdowntest

import (
	"Cadence-Cloud-Todo/markdown"
	"testing"
)

func TestParseTaskLine(t *testing.T) {
	cases := []struct {
		line string
		ok bool
		indent int
		status string
	}{
		{"- [ ] open", true, 0, "todo"},
		{"- [x] done", true, 0, "done"},
		{"- [X] done upper", true, 0, "done"},
		{"- [/] working", true, 0, "in_progress"},
		{"    - [ ] nested", true, 4, "todo"},
		{"- [?] unknown mark", false, 0, ""},
		{"-[ ] no space", false, 0, ""},
		{"- [ ]no space after", false, 0, ""},
		{"* [ ] wrong bullet", false, 0, ""},
		{"plain text", false, 0, ""},
		{"- [ ]", false, 0, ""},
		{"", false, 0, ""},
	}
	for _, tc := range cases {
		indent, task, ok := markdown.ParseTaskLine(tc.line)
		if ok != tc.ok {
			t.Errorf("%q: ok = %v, want %v", tc.line, ok, tc.ok)
			continue
		}
		if ok && (indent != tc.indent || task.Status != tc.status) {
			t.Errorf("%q: indent=%d status=%q", tc.line, indent, task.Status)
		}
	}
}