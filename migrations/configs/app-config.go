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

package configs

import (
	"github.com/joho/godotenv"
	"time"
	"os"
)

func EnvLoad() error {
	if err := godotenv.Load(".env"); err == nil {
		return nil
	}

	if err := godotenv.Load("../.env"); err != nil {
		return nil
	}

	return nil
}

func GetDBUrl() string {
	DB_URL := os.Getenv("MIGRATION_DB_URL")
	return DB_URL
}

func WatchInterval() time.Duration {
	return 10 * time.Second
}
