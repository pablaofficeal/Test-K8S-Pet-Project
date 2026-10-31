package models

import (
	"time"
)

type BotUsage struct {
	BotUUID string `json:"bot_uuid" gorm:"type:uuid;primaryKey"`
	Day time.Time `json:"day" gorm:"type:date;primaryKey"`
	OwnerUUID string `json:"owner_uuid" gorm:"type:uuid;not null;index:idx_bot_usage_owner_day,priority:1"`
	Messages int `json:"messages" gorm:"not null;default:0"`
}

func (BotUsage) TableName() string {
	return "bot_usage"
}