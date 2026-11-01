package models

import (
	"encoding/json"
	"time"
)

type TodoTask struct {
	ID string `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	BoardID string `json:"board_id" gorm:"type:uuid;not null;index"`
	NodeID string `json:"node_id" gorm:"type:uuid;not null;index"`
	ParentID *string `json:"parent_id" gorm:"type:uuid;index"`
	Title string `json:"title" gorm:"not null;default:''"`
	Description string `json:"description" gorm:"not null;default:''"`
	Status string `json:"status" gorm:"not null;default:todo"`
	Priority string `json:"priority" gorm:"not null;default:normal"`
	Tags json.RawMessage `json:"tags" gorm:"type:jsonb;not null;default:'[]'"`
	Deadline *time.Time `json:"deadline" gorm:"type:date"`
	SortOrder int `json:"sort_order" gorm:"not null;default:0"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}