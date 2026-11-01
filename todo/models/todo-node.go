package models

import "time"

type TodoNode struct {
	ID string `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	BoardID string `json:"board_id" gorm:"type:uuid;not null;index"`
	Title string `json:"title" gorm:"not null;default:''"`
	SortOrder int `json:"sort_order" gorm:"not null;default:0"`
	PosX float64 `json:"pos_x" gorm:"not null;default:0"`
	PosY float64 `json:"pos_y" gorm:"not null;default:0"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}