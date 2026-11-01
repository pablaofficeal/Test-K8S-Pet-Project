package cachetest

import (
	"Cadence-Cloud-Todo/cache"
	"context"
	"testing"
	"time"
)

func TestSetAndGetOrder(t *testing.T) {
	ctx := context.Background()
	redisServer.FlushAll()
	if _, found, _ := cache.GetOrder(ctx, "b1"); found {
		t.Error("a board without a hot order must report not found")
	}
	if err := cache.SetOrder(ctx, "b1", []string{"a", "b", "c"}); err != nil {
		t.Fatal(err)
	}
	cache.SetOrder(ctx, "b1", []string{"c", "a", "b"})
	got, found, err := cache.GetOrder(ctx, "b1")
	if err != nil || !found || len(got) != 3 || got[0] != "c" || got[1] != "a" || got[2] != "b" {
		t.Errorf("order = %v found=%v err=%v", got, found, err)
	}
	if _, found, _ := cache.GetOrder(ctx, "other"); found {
		t.Error("order of another board leaked")
	}
}

func TestSetOrderMarksBoardDirtyAndSetsTTL(t *testing.T) {
	ctx := context.Background()
	redisServer.FlushAll()
	cache.SetOrder(ctx, "b1", []string{"a"})
	members, _ := redisServer.SMembers("todo:dirty")
	if len(members) != 1 || members[0] != "b1" {
		t.Errorf("dirty set = %v", members)
	}
	if ttl := redisServer.TTL("todo:order:b1"); ttl <= 0 || ttl > 24*time.Hour {
		t.Errorf("ttl = %v", ttl)
	}
}

func TestGetOrderIgnoresBrokenValue(t *testing.T) {
	redisServer.FlushAll()
	redisServer.Set("todo:order:b1", "not json")
	if _, found, err := cache.GetOrder(context.Background(), "b1"); found || err != nil {
		t.Errorf("found=%v err=%v", found, err)
	}
}

func TestAppendAndRemoveKeepHotOrder(t *testing.T) {
	ctx := context.Background()
	redisServer.FlushAll()
	if err := cache.AppendToOrder(ctx, "b1", "x"); err != nil {
		t.Fatal(err)
	}
	if _, found, _ := cache.GetOrder(ctx, "b1"); found {
		t.Error("append must not create a hot order that does not exist")
	}
	cache.SetOrder(ctx, "b1", []string{"a", "b"})
	cache.AppendToOrder(ctx, "b1", "c")
	cache.RemoveFromOrder(ctx, "b1", "a")
	got, _, _ := cache.GetOrder(ctx, "b1")
	if len(got) != 2 || got[0] != "b" || got[1] != "c" {
		t.Errorf("order = %v", got)
	}
	if err := cache.RemoveFromOrder(ctx, "none", "a"); err != nil {
		t.Errorf("remove from a missing order: %v", err)
	}
}