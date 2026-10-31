/*
 * Copyright (c) 2026-2027 Analog WakaTime. All rights reserved.
 *
 * This file is part of Analog WakaTime Migration Service.
 *
 * This software is proprietary and confidential.
 * Unauthorized copying, modification, distribution, disclosure,
 * sublicensing, or use of this software, in whole or in part,
 * is strictly prohibited without prior written permission
 * from the copyright holder.
 */

package core

import (
	"Analog-WakaTime-Migration-Service/models"
	"gorm.io/gorm"
	"fmt"
)

func Seed(conn *gorm.DB) error {
	roles := []models.Roles{
		{RoleID: uint(models.RoleAdmin), Name: "admin"},
		{RoleID: uint(models.RoleUser), Name: "user"},
		{RoleID: uint(models.RolePartner), Name: "partner"},
		{RoleID: uint(models.RoleSupport), Name: "support"},
		{RoleID: uint(models.RoleUserOwner), Name: "user_owner"},
		{RoleID: uint(models.RoleAccountant), Name: "accountant"},
		{RoleID: uint(models.RoleModerator), Name: "moderator"},
	}
	for _, r := range roles {
		if err := conn.FirstOrCreate(&r, models.Roles{RoleID: r.RoleID}).Error; err != nil {
			return fmt.Errorf("seed role: %w", err)
		}
	}
	free := models.Plan{Key: "free", Name: "Free", MaxBots: 3, MemoryMB: 200, CPUMillis: 350, Pids: 64, DailyMessages: 200}
	if err := conn.FirstOrCreate(&free, models.Plan{Key: free.Key}).Error; err != nil {
		return fmt.Errorf("seed plan: %w", err)
	}
	return nil
}
