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

import "time"

type MessengerKey struct {
	UserUUID string `json:"user_uuid" gorm:"type:uuid;primaryKey"`
	KeyID string `json:"key_id" gorm:"not null;uniqueIndex"`
	PublicKey string `json:"public_key" gorm:"type:jsonb;not null"`
	Backup *string `json:"-" gorm:"type:text"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

type MessengerMessage struct {
	ID int64 `json:"id" gorm:"primaryKey;autoIncrement"`
	SenderUUID string `json:"sender_uuid" gorm:"type:uuid;not null;index:idx_messenger_pair,priority:1"`
	RecipientUUID string `json:"recipient_uuid" gorm:"type:uuid;not null;index:idx_messenger_pair,priority:2;index:idx_messenger_inbox,priority:1"`
	Envelope string `json:"envelope" gorm:"type:jsonb;not null"`
	ReadAt *time.Time `json:"read_at" gorm:"index:idx_messenger_inbox,priority:2"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
}