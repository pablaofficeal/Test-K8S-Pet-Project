package markdown

import "time"

type Task struct {
	Title string
	Description string
	Status string
	Priority string
	Tags []string
	Deadline *time.Time
	Children []*Task
}

type Node struct {
	Title string
	Tasks []*Task
}

type Document struct {
	Title string
	Nodes []*Node
}