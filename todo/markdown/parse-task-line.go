package markdown

import "strings"

func ParseTaskLine(line string) (int, *Task, bool) {
	trimmed := strings.TrimLeft(line, " ")
	indent := len(line) - len(trimmed)
	if len(trimmed) < 6 || trimmed[0] != '-' || trimmed[1] != ' ' || trimmed[2] != '[' || trimmed[4] != ']' || trimmed[5] != ' ' {
		return 0, nil, false
	}
	status := ""
	switch trimmed[3] {
	case ' ':
		status = "todo"
	case '/':
		status = "in_progress"
	case 'x', 'X':
		status = "done"
	default:
		return 0, nil, false
	}
	title, priority, tags, deadline := SplitMeta(trimmed[6:])
	return indent, &Task{Title: title, Status: status, Priority: priority, Tags: tags, Deadline: deadline}, true
}