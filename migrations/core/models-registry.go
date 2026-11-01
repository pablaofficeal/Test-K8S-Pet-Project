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

package core

import (
	"Cadence-Cloud-Migrations/models"
)

func Targets() []any {
	return []any{
		&models.User{},
		&models.Roles{},
		&models.UserRoles{},
		&models.RefreshToken{},
		&models.EmailVerification{},
		&models.TodoBoard{},
		&models.TodoNode{},
		&models.TodoTask{},
		&models.MessengerKey{},
		&models.MessengerMessage{},
	}
}
