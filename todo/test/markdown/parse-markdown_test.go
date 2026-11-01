package markdowntest

import (
	"Cadence-Cloud-Todo/markdown"
	"testing"
	"os"
)

func TestParseSample(t *testing.T) {
	raw, err := os.ReadFile("testdata/sample.md")
	if err != nil {
		t.Fatal(err)
	}
	doc, err := markdown.Parse(string(raw))
	if err != nil {
		t.Fatal(err)
	}
	if doc.Title != "PureC-OS todo" {
		t.Errorf("title = %q", doc.Title)
	}
	titles := []string{}
	for _, n := range doc.Nodes {
		titles = append(titles, n.Title)
	}
	want := []string{"UserSpace", "Userspace programs", "Network", "Kernel", "Touchpad", "SMP и Много процесорность", "SO формат"}
	if len(titles) != len(want) {
		t.Fatalf("nodes = %v, want %v", titles, want)
	}
	for i := range want {
		if titles[i] != want[i] {
			t.Errorf("node %d = %q, want %q", i, titles[i], want[i])
		}
	}

	bug := doc.Nodes[0].Tasks[4]
	if bug.Title != "User space ведёт себя не так как задумано" || bug.Status != "done" || bug.Priority != "high" {
		t.Errorf("bug task parsed wrong: %+v", bug)
	}
	if len(bug.Tags) != 1 || bug.Tags[0] != "bug" || bug.Deadline == nil || bug.Deadline.Format("2006-01-02") != "2026-09-23" {
		t.Errorf("bug metadata parsed wrong: tags=%v deadline=%v", bug.Tags, bug.Deadline)
	}
	wantDesc := "Когда userspace а точнее window manager стартует в отдельном процессе Я открываю окно что-то делаю в программе потом закрываю окно\nИ полностью вся система падает с какой ошибкой я не знаю\nПросто зависает"
	if bug.Description != wantDesc {
		t.Errorf("multi-line description = %q", bug.Description)
	}

	wifi := doc.Nodes[2].Tasks[0]
	if wifi.Description != "Частично было добавлено" || len(wifi.Children) != 1 || wifi.Children[0].Title != "Добавить поддержку 802.11b" {
		t.Errorf("description followed by a subtask parsed wrong: %+v", wifi)
	}
	glued := doc.Nodes[2].Tasks[2]
	if glued.Title != "Поднять Atheros AR928X Wirless Network Adapter (Этот пункт частично реализован, но не работаеть)" || glued.Priority != "normal" {
		t.Errorf("task with glued pipe parsed wrong: %q", glued.Title)
	}
	deep := doc.Nodes[3].Tasks[1].Children[0].Children[0]
	if deep.Title != "Скорее всего будет через vender soft" {
		t.Errorf("third level task = %q", deep.Title)
	}
	nested := doc.Nodes[5].Tasks[0].Children[0]
	if nested.Description == "" || nested.Children != nil {
		t.Errorf("description of a nested task lost: %+v", nested)
	}
	if doc.Nodes[6].Tasks[0].Status != "in_progress" {
		t.Errorf("[/] must be in_progress, got %q", doc.Nodes[6].Tasks[0].Status)
	}
}

func TestParseRejectsEmptyDocument(t *testing.T) {
	for _, in := range []string{"", "\n\n", "# Only a title\n", "just text"} {
		if _, err := markdown.Parse(in); err == nil {
			t.Errorf("Parse(%q) must fail", in)
		}
	}
}

func TestParseTasksBeforeFirstSection(t *testing.T) {
	doc, err := markdown.Parse("- [ ] loose task | priority:low\n## A\n- [x] done")
	if err != nil {
		t.Fatal(err)
	}
	if len(doc.Nodes) != 2 || doc.Nodes[0].Title != "" || doc.Nodes[0].Tasks[0].Priority != "low" {
		t.Errorf("unexpected structure: %+v", doc.Nodes)
	}
}

func TestParseWindowsLineEndings(t *testing.T) {
	doc, err := markdown.Parse("# T\r\n\r\n## A\r\n\r\n- [ ] one\r\n  note\r\n")
	if err != nil {
		t.Fatal(err)
	}
	if doc.Nodes[0].Tasks[0].Title != "one" || doc.Nodes[0].Tasks[0].Description != "note" {
		t.Errorf("CRLF parsed wrong: %+v", doc.Nodes[0].Tasks[0])
	}
}

func TestParseRejectsTooDeepNesting(t *testing.T) {
	source := "## A\n"
	for i := 0; i < markdown.MaxDepth+1; i++ {
		pad := ""
		for j := 0; j < i; j++ {
			pad += "  "
		}
		source += pad + "- [ ] level\n"
	}
	if _, err := markdown.Parse(source); err == nil {
		t.Error("expected nesting error")
	}
}