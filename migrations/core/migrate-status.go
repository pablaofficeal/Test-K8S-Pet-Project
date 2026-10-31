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
	"gorm.io/gorm"
	"fmt"
)

func Status(conn *gorm.DB) error {
	for _, m := range Targets() {
		fmt.Printf("%T ready=%v\n", m, conn.Migrator().HasTable(m))
	}
	return nil
}
