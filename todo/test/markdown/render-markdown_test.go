package markdowntest

import (
	"Cadence-Cloud-Todo/markdown"
	"reflect"
	"testing"
	"os"
)

func TestRoundTripKeepsStructure(t *testing.T) {
	raw, _ := os.ReadFile("testdata/sample.md")
	first, err := markdown.Parse(string(raw))
	if err != nil {
		t.Fatal(err)
	}
	rendered := markdown.Render(first)
	second, err := markdown.Parse(rendered)
	if err != nil {
		t.Fatalf("rendered markdown does not parse: %v", err)
	}
	if !reflect.DeepEqual(first, second) {
		t.Error("parse(render(parse(x))) differs from parse(x)")
	}
	if markdown.Render(second) != rendered {
		t.Error("render is not stable on a second pass")
	}
}

func TestRenderFormat(t *testing.T) {
	doc, _ := markdown.Parse("# Plan\n\n## Work\n\n- [/] ship it | priority:high | tags:a,b | deadline:2026-01-02\n  line one\n  line two\n  - [x] sub | priority:low\n")
	want := "# Plan\n\n## Work\n\n- [/] ship it | priority:high | tags:a,b | deadline:2026-01-02\n  line one\n  line two\n  - [x] sub | priority:low"
	if got := markdown.Render(doc); got != want {
		t.Errorf("render mismatch:\n%s\n--- want ---\n%s", got, want)
	}
}

func TestRenderOmitsEmptyOptionalMeta(t *testing.T) {
	got := markdown.Render(&markdown.Document{Nodes: []*markdown.Node{{Title: "N", Tasks: []*markdown.Task{{Title: "t", Status: "todo"}}}}})
	if got != "## N\n\n- [ ] t | priority:normal" {
		t.Errorf("unexpected render: %q", got)
	}
}