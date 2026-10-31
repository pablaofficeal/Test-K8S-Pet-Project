package models

import "time"

type RefreshToken struct {
	UserUUID string `gorm:"not null;index" json:"user_uuid"`
	SID string `gorm:"not null;uniqueIndex" json:"sid"`
	TokenHash string `gorm:"not null" json:"-"`
	ExpiresAt time.Time `gorm:"not null;index" json:"expires_at"`
	Revoked bool `gorm:"not null;default:false" json:"revoked"`
	RevokedAt *time.Time `json:"revoked_at,omitempty"`
	CreatedAt time.Time `gorm:"not null" json:"created_at"`
	UpdatedAt time.Time `gorm:"not null" json:"updated_at"`
}