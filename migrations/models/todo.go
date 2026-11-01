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

type TodoBoard struct {
	ID string `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	OwnerUUID string `json:"owner_uuid" gorm:"type:uuid;not null;index"`
	Title string `json:"title" gorm:"not null;default:''"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

type TodoNode struct {
	ID string `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	BoardID string `json:"board_id" gorm:"type:uuid;not null;index"`
	Title string `json:"title" gorm:"not null;default:''"`
	SortOrder int `json:"sort_order" gorm:"not null;default:0"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

type TodoTask struct {
	ID string `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	BoardID string `json:"board_id" gorm:"type:uuid;not null;index"`
	NodeID string `json:"node_id" gorm:"type:uuid;not null;index"`
	ParentID *string `json:"parent_id" gorm:"type:uuid;index"`
	Title string `json:"title" gorm:"not null;default:''"`
	Description string `json:"description" gorm:"not null;default:''"`
	Status string `json:"status" gorm:"not null;default:todo"`
	Priority string `json:"priority" gorm:"not null;default:normal"`
	Tags string `json:"tags" gorm:"type:jsonb;not null;default:'[]'"`
	Deadline *time.Time `json:"deadline" gorm:"type:date"`
	SortOrder int `json:"sort_order" gorm:"not null;default:0"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}