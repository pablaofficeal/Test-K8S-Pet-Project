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

func Down(conn *gorm.DB) error {
	targets := Targets()
	for i := len(targets) - 1; i >= 0; i-- {
		if err := conn.Migrator().DropTable(targets[i]); err != nil {
			return fmt.Errorf("drop table: %w", err)
		}
	}
	return nil
}
