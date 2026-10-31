package models

import (
	"time"
)

type Plan struct {
	Key string `json:"key" gorm:"primaryKey"`
	Name string `json:"name" gorm:"not null"`
	MaxBots int `json:"max_bots" gorm:"column:max_bots;not null"`
	MemoryMB int `json:"memory_mb" gorm:"column:memory_mb;not null"`
	CPUMillis int `json:"cpu_millis" gorm:"column:cpu_millis;not null"`
	Pids int `json:"pids" gorm:"column:pids;not null"`
	DailyMessages int `json:"daily_messages" gorm:"column:daily_messages;not null"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

type UserLimit struct {
	UserUUID string `json:"user_uuid" gorm:"type:uuid;primaryKey"`
	PlanKey *string `json:"plan_key" gorm:"column:plan_key"`
	MaxBots *int `json:"max_bots" gorm:"column:max_bots"`
	MemoryMB *int `json:"memory_mb" gorm:"column:memory_mb"`
	CPUMillis *int `json:"cpu_millis" gorm:"column:cpu_millis"`
	Pids *int `json:"pids" gorm:"column:pids"`
	DailyMessages *int `json:"daily_messages" gorm:"column:daily_messages"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}
