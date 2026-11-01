package cachetest

import (
	"Cadence-Cloud-Todo/cache"
	"context"
	"testing"
	"time"
)

func TestSetAndGetPositions(t *testing.T) {
	ctx := context.Background()
	redisServer.FlushAll()
	if err := cache.SetPosition(ctx, "b1", "n1", 10.5, -20); err != nil {
		t.Fatal(err)
	}
	cache.SetPosition(ctx, "b1", "n2", 3, 4)
	cache.SetPosition(ctx, "b1", "n1", 99, 100)

	got, err := cache.GetPositions(ctx, "b1")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got["n1"] != [2]float64{99, 100} || got["n2"] != [2]float64{3, 4} {
		t.Errorf("positions = %v", got)
	}
	if other, _ := cache.GetPositions(ctx, "other"); len(other) != 0 {
		t.Errorf("positions of another board leaked: %v", other)
	}
}

func TestSetPositionMarksBoardDirtyAndSetsTTL(t *testing.T) {
	ctx := context.Background()
	redisServer.FlushAll()
	cache.SetPosition(ctx, "b1", "n1", 1, 2)
	members, _ := redisServer.SMembers("todo:dirty")
	if len(members) != 1 || members[0] != "b1" {
		t.Errorf("dirty set = %v", members)
	}
	if ttl := redisServer.TTL("todo:pos:b1"); ttl <= 0 || ttl > 24*time.Hour {
		t.Errorf("ttl = %v", ttl)
	}
}

func TestGetPositionsSkipsBrokenValues(t *testing.T) {
	ctx := context.Background()
	redisServer.FlushAll()
	redisServer.HSet("todo:pos:b1", "good", "1:2", "nocolon", "12", "nan", "a:b")
	got, _ := cache.GetPositions(ctx, "b1")
	if len(got) != 1 || got["good"] != [2]float64{1, 2} {
		t.Errorf("positions = %v", got)
	}
}

func TestDeletePosition(t *testing.T) {
	ctx := context.Background()
	redisServer.FlushAll()
	cache.SetPosition(ctx, "b1", "n1", 1, 2)
	cache.SetPosition(ctx, "b1", "n2", 3, 4)
	redisServer.Set("todo:node:n1", "owner:b1")
	if err := cache.DeletePosition(ctx, "b1", "n1"); err != nil {
		t.Fatal(err)
	}
	got, _ := cache.GetPositions(ctx, "b1")
	if _, ok := got["n1"]; ok || len(got) != 1 {
		t.Errorf("positions after delete = %v", got)
	}
	if redisServer.Exists("todo:node:n1") {
		t.Error("owner cache of the deleted node must be dropped")
	}
}