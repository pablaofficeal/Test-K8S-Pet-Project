package models

import "time"

type EmailVerification struct {
	ID         uint      `gorm:"primaryKey"`
	UserUUID   string    `gorm:"type:uuid;index;not null"`
	Email      string    `gorm:"not null"`
	CodeHash   string    `gorm:"not null"`
	Attempts   int       `gorm:"not null;default:0"`
	ExpiresAt  time.Time `gorm:"not null;index"`
	ConsumedAt *time.Time
	CreatedAt  time.Time `gorm:"autoCreateTime"`
}