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
	"Cadence-Cloud-Migrations/configs"
	"gorm.io/gorm"
	"strings"
	"bufio"
	"fmt"
	"os"
)

var stdin = bufio.NewReader(os.Stdin)

func readLine() (string, bool) {
	line, err := stdin.ReadString('\n')
	if err != nil {
		return "", false
	}
	return strings.TrimSpace(line), true
}

func resolveTargets(name string, targets []configs.DBTarget) ([]configs.DBTarget, error) {
	if name == "" || name == "all" {
		return targets, nil
	}
	for _, t := range targets {
		if t.Name == name {
			return []configs.DBTarget{t}, nil
		}
	}
	return nil, fmt.Errorf("unknown database: %s", name)
}

func applyAction(conn *gorm.DB, action string) error {
	switch action {
	case "up":
		if err := Up(conn); err != nil {
			return err
		}
		return Seed(conn)
	case "down":
		return Down(conn)
	case "status":
		return Status(conn)
	case "seed":
		return Seed(conn)
	}
	return fmt.Errorf("unknown action: %s", action)
}

func pickTarget(targets []configs.DBTarget) ([]configs.DBTarget, bool) {
	fmt.Println("Databases:")
	for i, t := range targets {
		fmt.Printf("  %d) %s\n", i+1, t.Name)
	}
	fmt.Println("  0) all")
	fmt.Println("  q) exit")
	fmt.Print("Select database: ")
	line, ok := readLine()
	if !ok || line == "q" || line == "exit" {
		return nil, false
	}
	if line == "0" || line == "all" {
		return targets, true
	}
	for i, t := range targets {
		if line == fmt.Sprint(i+1) || line == t.Name {
			return []configs.DBTarget{t}, true
		}
	}
	fmt.Println("Unknown choice")
	return pickTarget(targets)
}

func pickAction() (string, bool) {
	fmt.Println("Actions:")
	fmt.Println("  1) up")
	fmt.Println("  2) down")
	fmt.Println("  3) status")
	fmt.Println("  4) seed")
	fmt.Println("  5) watch")
	fmt.Println("  q) exit")
	fmt.Print("Select action: ")
	line, ok := readLine()
	if !ok {
		return "", false
	}
	switch line {
	case "1", "up":
		return "up", true
	case "2", "down":
		return "down", true
	case "3", "status":
		return "status", true
	case "4", "seed":
		return "seed", true
	case "5", "watch":
		return "watch", true
	}
	return "", false
}

func askConfirm(picked []configs.DBTarget) bool {
	fmt.Print("Drop tables on [")
	for i, t := range picked {
		if i > 0 {
			fmt.Print(", ")
		}
		fmt.Print(t.Name)
	}
	fmt.Print("]? Type yes: ")
	line, ok := readLine()
	return ok && line == "yes"
}
