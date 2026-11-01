/*
 * Copyright (c) 2026-2027 Cadence Cloud. All rights reserved.
 *
 * This file is part of Cadence Cloud Migration Service.
 *
 * This software is proprietary and confidential.
 * Unauthorized copying, modification, distribution, disclosure,
 * sublicensing, or use of this software, in whole or in part,
 * is strictly prohibited without prior written permission
 * from the copyright holder.
 */

package models

import (
	"time"
)

type User struct {
	UUID string `gorm:"type:uuid;default:gen_random_uuid();uniqueIndex"`
	Username string `gorm:"type:string;nullable:false"`
	Email string `gorm:"type:string;nullable:false"`
	Password string `json:"-" gorm:"not null"`
	IsActive bool `json:"is_active" gorm:"not null;default:true"`
	EmailVerified bool `json:"email_verified" gorm:"not null;default:false"`
	EmailVerifiedAt *time.Time `json:"email_verified_at"`
	AvatarKey *string `json:"avatar_key" gorm:"type:text"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}