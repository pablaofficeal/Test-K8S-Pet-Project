package models

import "time"

type User struct {
	UUID string `gorm:"type:uuid;default:gen_random_uuid();uniqueIndex"`
	Username string `gorm:"type:string;nullable:false"`
	Email string `gorm:"type:string;nullable:false"`
	Password string `json:"-" gorm:"not null"`
	IsActive bool `json:"is_active" gorm:"not null;default:true"`
	EmailVerified bool `json:"email_verified" gorm:"not null;default:false"`
	EmailVerifiedAt *time.Time `json:"email_verified_at"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}