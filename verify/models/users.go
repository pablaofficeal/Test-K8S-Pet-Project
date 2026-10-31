package models

import "time"

type User struct {
	UUID            string `gorm:"type:uuid;uniqueIndex"`
	Username        string
	Email           string
	EmailVerified   bool `gorm:"not null;default:false"`
	EmailVerifiedAt *time.Time
}