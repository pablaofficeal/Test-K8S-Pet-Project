package db

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"time"
	"log"
	"fmt"
)

var DB *gorm.DB

func InitDB(dsn string) error {
	var err error

	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("connect to database: %w", err)
	}

	return nil
}

func WaitForDB(dsn string, interval time.Duration) {
	for {
		if err := InitDB(dsn); err == nil {
			sqlDB, err := DB.DB()
			if err == nil {
				if err := sqlDB.Ping(); err == nil {
					return
				}
			}
		} else {
			log.Printf("postgres not ready, retry in %s: %v", interval.String(), err)
		}
		time.Sleep(interval)
	}
}

func WaitForTables(interval time.Duration, targets ...any) {
	for {
		ready := DB != nil
		if ready {
			for _, t := range targets {
				if !DB.Migrator().HasTable(t) {
					ready = false
					break
				}
			}
		}
		if ready {
			return
		}
		log.Printf("tables not ready, retry in %s", interval.String())
		time.Sleep(interval)
	}
}