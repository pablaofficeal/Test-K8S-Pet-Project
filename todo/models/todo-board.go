package models

import "time"

type TodoBoard struct {
	ID string `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	OwnerUUID string `json:"owner_uuid" gorm:"type:uuid;not null;index"`
	Title string `json:"title" gorm:"not null;default:''"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}