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
	"Analog-WakaTime-Migration-Service/configs"
	"Analog-WakaTime-Migration-Service/db"
	"gorm.io/gorm"
	"time"
	"log"
)

func WatchAll(targets []configs.DBTarget, interval time.Duration) {
	conns := make(map[string]*gorm.DB, len(targets))
	for _, t := range targets {
		conns[t.Name] = db.WaitForDB(t.Name, t.URL, interval)
	}
	for {
		for _, t := range targets {
			if err := Up(conns[t.Name]); err != nil {
				log.Printf("[%s] migrate up failed: %v", t.Name, err)
				continue
			}
			if err := Seed(conns[t.Name]); err != nil {
				log.Printf("[%s] seed failed: %v", t.Name, err)
			} else {
				log.Printf("[%s] migrate up ok", t.Name)
			}
		}
		time.Sleep(interval)
	}
}
