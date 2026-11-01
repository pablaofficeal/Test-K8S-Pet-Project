package markdown

import (
	"strings"
)

func Render(doc *Document) string {
	var out strings.Builder
	if doc.Title != "" {
		out.WriteString("# " + doc.Title + "\n\n")
	}
	var write func(tasks []*Task, depth int)
	write = func(tasks []*Task, depth int) {
		pad := strings.Repeat("  ", depth)
		for _, t := range tasks {
			mark := "[ ]"
			if t.Status == "in_progress" {
				mark = "[/]"
			} else if t.Status == "done" {
				mark = "[x]"
			}
			priority := t.Priority
			if priority == "" {
				priority = "normal"
			}
			out.WriteString(pad + "- " + mark + " " + t.Title + " | priority:" + priority)
			if len(t.Tags) > 0 {
				out.WriteString(" | tags:" + strings.Join(t.Tags, ","))
			}
			if t.Deadline != nil {
				out.WriteString(" | deadline:" + t.Deadline.Format("2006-01-02"))
			}
			out.WriteString("\n")
			if t.Description != "" {
				for _, line := range strings.Split(t.Description, "\n") {
					out.WriteString(pad + "  " + line + "\n")
				}
			}
			write(t.Children, depth+1)
		}
	}
	for i, node := range doc.Nodes {
		if node.Title != "" {
			out.WriteString("## " + node.Title + "\n\n")
		}
		write(node.Tasks, 0)
		if i < len(doc.Nodes)-1 {
			out.WriteString("\n")
		}
	}
	return strings.TrimRight(out.String(), "\n")
}