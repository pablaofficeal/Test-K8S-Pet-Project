package markdown

import (
	"strings"
	"fmt"
)

const MaxDepth = 8

func Parse(source string) (*Document, error) {
	doc := &Document{}
	type frame struct {
		indent int
		task *Task
	}
	var stack []frame
	var current *Node
	for _, raw := range strings.Split(strings.ReplaceAll(source, "\r\n", "\n"), "\n") {
		line := strings.TrimRight(raw, " \t")
		if strings.TrimSpace(line) == "" {
			continue
		}
		if strings.HasPrefix(line, "## ") {
			current = &Node{Title: strings.TrimSpace(line[3:])}
			doc.Nodes = append(doc.Nodes, current)
			stack = nil
			continue
		}
		if strings.HasPrefix(line, "# ") && doc.Title == "" && len(doc.Nodes) == 0 {
			doc.Title = strings.TrimSpace(line[2:])
			continue
		}
		if indent, task, ok := ParseTaskLine(line); ok {
			if current == nil {
				current = &Node{}
				doc.Nodes = append(doc.Nodes, current)
			}
			for len(stack) > 0 && stack[len(stack)-1].indent >= indent {
				stack = stack[:len(stack)-1]
			}
			if len(stack) >= MaxDepth {
				return nil, fmt.Errorf("tasks are nested deeper than %d levels", MaxDepth)
			}
			if len(stack) == 0 {
				current.Tasks = append(current.Tasks, task)
			} else {
				parent := stack[len(stack)-1].task
				parent.Children = append(parent.Children, task)
			}
			stack = append(stack, frame{indent: indent, task: task})
			continue
		}
		if len(stack) > 0 {
			top := stack[len(stack)-1]
			lead := len(line) - len(strings.TrimLeft(line, " "))
			if lead > top.indent {
				text := line[min(lead, top.indent+2):]
				if top.task.Description != "" {
					top.task.Description += "\n"
				}
				top.task.Description += text
			}
		}
	}
	if len(doc.Nodes) == 0 {
		return nil, fmt.Errorf("document has no sections or tasks")
	}
	return doc, nil
}