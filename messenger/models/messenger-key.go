package models

import "time"

type MessengerKey struct {
	UserUUID string `json:"user_uuid" gorm:"type:uuid;primaryKey"`
	KeyID string `json:"key_id" gorm:"not null;uniqueIndex"`
	PublicKey string `json:"public_key" gorm:"type:jsonb;not null"`
	Backup *string `json:"-" gorm:"type:text"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}
