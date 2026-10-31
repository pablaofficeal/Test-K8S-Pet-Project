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
)

func Targets() []any {
	return []any{
		&models.User{},
		&models.Roles{},
		&models.UserRoles{},
		&models.RefreshToken{},
		&models.EmailVerification{},
		&models.Bot{},
		&models.Plan{},
		&models.UserLimit{},
		&models.BotUsage{},
	}
}
