package main

import (
	"Cadence-Cloud-Migrations/configs"
	"Cadence-Cloud-Migrations/core"
	"Cadence-Cloud-Migrations/db"
	"log"
	"fmt"
	"os"
)

func main() {
	os.MkdirAll("logs", 0755)
	f, _ := os.OpenFile("logs/migration-service.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	log.SetOutput(f)
	db.LogOut = f
	if err := configs.EnvLoad(); err != nil {
		log.Fatal(err)
	}
	targets := configs.GetDBTargets()
	if len(targets) == 0 {
		fmt.Println("no DATABASE_URL entries found")
		log.Fatal("no database targets")
	}
	interval := configs.WatchInterval()
	if len(os.Args) > 1 {
		name := ""
		if len(os.Args) > 2 {
			name = os.Args[2]
		}
		if err := core.RunCommand(os.Args[1], name, targets, interval); err != nil {
			fmt.Println("FAILED:", err)
			log.Printf("command %s failed: %v", os.Args[1], err)
			os.Exit(1)
		}
		return
	}
	core.RunInteractive(targets, interval)
}
