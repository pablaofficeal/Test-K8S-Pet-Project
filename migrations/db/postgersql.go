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

package db

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm/logger"
	"gorm.io/gorm"
	"time"
	"log"
	"fmt"
	"io"
)

var LogOut io.Writer = io.Discard

func newLogger() logger.Interface {
	return logger.New(log.New(LogOut, "", log.LstdFlags), logger.Config{
		SlowThreshold: 200 * time.Millisecond,
		LogLevel:      logger.Warn,
		Colorful:      false,
	})
}

func Connect(url string) (*gorm.DB, error) {
	conn, err := gorm.Open(postgres.Open(url), &gorm.Config{Logger: newLogger()})
	if err != nil {
		return nil, fmt.Errorf("connect to database: %w", err)
	}
	sqlDB, err := conn.DB()
	if err != nil {
		return nil, fmt.Errorf("get sql db: %w", err)
	}
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}
	return conn, nil
}

func WaitForDB(name string, url string, interval time.Duration) *gorm.DB {
	for {
		conn, err := Connect(url)
		if err == nil {
			return conn
		}
		log.Printf("[%s] postgres not ready, retry in %s: %v", name, interval.String(), err)
		time.Sleep(interval)
	}
}

func HasAllTables(conn *gorm.DB, targets []any) bool {
	for _, t := range targets {
		if !conn.Migrator().HasTable(t) {
			return false
		}
	}
	return true
}
