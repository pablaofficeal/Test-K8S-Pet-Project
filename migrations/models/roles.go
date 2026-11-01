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

type RoleID uint

const (
	RoleAdmin RoleID = 1
	RoleUser RoleID = 2
	RolePartner RoleID = 3
	RoleSupport RoleID = 4
	RoleUserOwner RoleID = 5
	RoleAccountant RoleID = 6
	RoleModerator RoleID = 7
)

type Roles struct {
	RoleID uint `gorm:"primaryKey;autoIncrement"`
	Name string `gorm:"not null" json:"name"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type UserRoles struct {
	UserUUID string `json:"user_uuid" gorm:"type:uuid;primaryKey"`
	RoleID RoleID `json:"role_id" gorm:"primaryKey"`
}