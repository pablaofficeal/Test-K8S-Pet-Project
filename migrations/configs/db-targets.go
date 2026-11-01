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
	"strings"
	"sort"
	"os"
)

type DBTarget struct {
	Name string
	URL  string
}

func GetDBTargets() []DBTarget {
	found := map[string]string{}
	for _, e := range os.Environ() {
		i := strings.Index(e, "=")
		if i < 0 {
			continue
		}
		key, value := e[:i], e[i+1:]
		name := ""
		if key == "MIGRATION_DB_URL" {
			name = "default"
		} else if strings.HasPrefix(key, "MIGRATION_DB_URL_") {
			name = strings.ToLower(strings.TrimPrefix(key, "MIGRATION_DB_URL_"))
		} else {
			continue
		}
		if value == "" {
			continue
		}
		found[name] = value
	}
	targets := make([]DBTarget, 0, len(found))
	for n, u := range found {
		targets = append(targets, DBTarget{Name: n, URL: u})
	}
	sort.Slice(targets, func(i int, j int) bool { return targets[i].Name < targets[j].Name })
	return targets
}
