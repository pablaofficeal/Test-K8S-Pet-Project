
package models

import (
	"time"
)

type BotStatus string

const (
	BotStatusStopped BotStatus = "stopped"
	BotStatusRunning BotStatus = "running"
	BotStatusError BotStatus = "error"
)

type Bot struct {
	UUID string `json:"uuid" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	OwnerUUID string `json:"owner_uuid" gorm:"type:uuid;not null;index;uniqueIndex:idx_bots_owner_name"`
	Name string `json:"name" gorm:"not null;uniqueIndex:idx_bots_owner_name"`
	Username string `json:"username" gorm:"not null;default:''"`
	Description string `json:"description" gorm:"not null;default:''"`
	WelcomeMessage string `json:"welcome_message" gorm:"not null;default:''"`
	AvatarKey string `json:"-" gorm:"not null;default:''"`
	TokenEncrypted string `json:"-" gorm:"not null"`
	Status BotStatus `json:"status" gorm:"not null;default:stopped;index"`
	TemplateID string `json:"template_id" gorm:"not null;default:''"`
	Library string `json:"library" gorm:"not null;default:''"`
	SourcePrefix string `json:"-" gorm:"not null;default:''"`
	StatusMessage string `json:"status_message" gorm:"not null;default:''"`
	ContainerID string `json:"container_id,omitempty"`
	Config string `json:"config" gorm:"type:jsonb;not null;default:'{}'"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}