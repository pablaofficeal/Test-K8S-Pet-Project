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
	"gorm.io/gorm"
	"fmt"
)

func Up(conn *gorm.DB) error {
	if err := conn.Exec(`CREATE EXTENSION IF NOT EXISTS "pgcrypto"`).Error; err != nil {
		return fmt.Errorf("enable pgcrypto: %w", err)
	}
	if err := conn.AutoMigrate(Targets()...); err != nil {
		return fmt.Errorf("automigrate: %w", err)
	}
	return nil
}
