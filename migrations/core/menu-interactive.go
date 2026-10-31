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
	"time"
	"log"
	"fmt"
)

func RunCommand(cmd string, name string, targets []configs.DBTarget, interval time.Duration) error {
	picked, err := resolveTargets(name, targets)
	if err != nil {
		return err
	}
	if cmd == "watch" {
		WatchAll(picked, interval)
		return nil
	}
	for _, t := range picked {
		conn, err := db.Connect(t.URL)
		if err != nil {
			return fmt.Errorf("[%s]: %w", t.Name, err)
		}
		log.Printf("[%s] run %s", t.Name, cmd)
		if err := applyAction(conn, cmd); err != nil {
			return fmt.Errorf("[%s]: %w", t.Name, err)
		}
		fmt.Printf("[%s] %s ok\n", t.Name, cmd)
	}
	return nil
}

func RunInteractive(targets []configs.DBTarget, interval time.Duration) {
	for {
		picked, ok := pickTarget(targets)
		if !ok {
			return
		}
		action, ok := pickAction()
		if !ok {
			return
		}
		if action == "down" && !askConfirm(picked) {
			continue
		}
		if action == "watch" {
			WatchAll(picked, interval)
			return
		}
		for _, t := range picked {
			conn, err := db.Connect(t.URL)
			if err != nil {
				fmt.Printf("[%s] connect failed: %v\n", t.Name, err)
				log.Printf("[%s] connect failed: %v", t.Name, err)
				continue
			}
			log.Printf("[%s] run %s", t.Name, action)
			if err := applyAction(conn, action); err != nil {
				fmt.Printf("[%s] %s failed: %v\n", t.Name, action, err)
				log.Printf("[%s] %s failed: %v", t.Name, action, err)
				continue
			}
			fmt.Printf("[%s] %s ok\n", t.Name, action)
		}
	}
}
