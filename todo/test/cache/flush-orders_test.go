package cachetest

import (
	"Cadence-Cloud-Todo/models"
	"Cadence-Cloud-Todo/cache"
	"Cadence-Cloud-Todo/db"
	"github.com/google/uuid"
	"context"
	"testing"
	"time"
)

func TestFlushWithoutDirtyBoards(t *testing.T) {
	redisServer.FlushAll()
	n, err := cache.FlushOrders(context.Background())
	if err != nil || n != 0 {
		t.Errorf("flush = %d, %v", n, err)
	}
}

func TestFlushOrdersWritesToPostgres(t *testing.T) {
	if !dbReady {
		t.Skip("postgres is not available")
	}
	ctx := context.Background()
	redisServer.FlushAll()
	original := db.DB
	tx := original.Begin()
	db.DB = tx
	t.Cleanup(func() { tx.Rollback(); db.DB = original })

	board := models.TodoBoard{OwnerUUID: uuid.NewString(), Title: "flush"}
	tx.Create(&board)
	a := models.TodoNode{BoardID: board.ID, Title: "a", SortOrder: 0}
	b := models.TodoNode{BoardID: board.ID, Title: "b", SortOrder: 1}
	c := models.TodoNode{BoardID: board.ID, Title: "c", SortOrder: 2}
	tx.Create(&a)
	tx.Create(&b)
	tx.Create(&c)

	cache.SetOrder(ctx, board.ID, []string{c.ID, a.ID, b.ID})
	var before models.TodoNode
	tx.First(&before, "id = ?", c.ID)
	if before.SortOrder != 2 {
		t.Fatalf("order reached postgres before the flush: %d", before.SortOrder)
	}

	n, err := cache.FlushOrders(ctx)
	if err != nil || n != 3 {
		t.Fatalf("flush = %d, %v", n, err)
	}
	for id, want := range map[string]int{c.ID: 0, a.ID: 1, b.ID: 2} {
		var got models.TodoNode
		tx.First(&got, "id = ?", id)
		if got.SortOrder != want {
			t.Errorf("node %s sort_order = %d, want %d", id, got.SortOrder, want)
		}
	}
	if members, _ := redisServer.SMembers("todo:dirty"); len(members) != 0 {
		t.Errorf("dirty set must be empty after the flush, got %v", members)
	}
	if _, found, _ := cache.GetOrder(ctx, board.ID); !found {
		t.Error("hot order must stay readable after the flush")
	}
}

func TestFlushIgnoresNodesOfOtherBoards(t *testing.T) {
	if !dbReady {
		t.Skip("postgres is not available")
	}
	ctx := context.Background()
	redisServer.FlushAll()
	original := db.DB
	tx := original.Begin()
	db.DB = tx
	t.Cleanup(func() { tx.Rollback(); db.DB = original })

	mine := models.TodoBoard{OwnerUUID: uuid.NewString(), Title: "mine"}
	theirs := models.TodoBoard{OwnerUUID: uuid.NewString(), Title: "theirs"}
	tx.Create(&mine)
	tx.Create(&theirs)
	foreign := models.TodoNode{BoardID: theirs.ID, Title: "foreign", SortOrder: 7}
	tx.Create(&foreign)

	cache.SetOrder(ctx, mine.ID, []string{foreign.ID})
	cache.FlushOrders(ctx)

	var got models.TodoNode
	tx.First(&got, "id = ?", foreign.ID)
	if got.SortOrder != 7 {
		t.Error("flush wrote an order into a node that belongs to another board")
	}
}

func TestStartFlusherFlushesPeriodicallyAndOnShutdown(t *testing.T) {
	if !dbReady {
		t.Skip("postgres is not available")
	}
	redisServer.FlushAll()
	original := db.DB
	tx := original.Begin()
	db.DB = tx
	t.Cleanup(func() { tx.Rollback(); db.DB = original })

	board := models.TodoBoard{OwnerUUID: uuid.NewString(), Title: "flusher"}
	tx.Create(&board)
	a := models.TodoNode{BoardID: board.ID, Title: "a", SortOrder: 0}
	b := models.TodoNode{BoardID: board.ID, Title: "b", SortOrder: 1}
	tx.Create(&a)
	tx.Create(&b)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	cache.StartFlusher(ctx, 30*time.Millisecond, done)
	cache.SetOrder(context.Background(), board.ID, []string{b.ID, a.ID})
	time.Sleep(300 * time.Millisecond)

	var got models.TodoNode
	tx.First(&got, "id = ?", b.ID)
	if got.SortOrder != 0 {
		t.Errorf("periodic flush did not happen: %d", got.SortOrder)
	}

	cache.SetOrder(context.Background(), board.ID, []string{a.ID, b.ID})
	cancel()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("flusher did not stop")
	}
	tx.First(&got, "id = ?", b.ID)
	if got.SortOrder != 1 {
		t.Errorf("final flush on shutdown did not happen: %d", got.SortOrder)
	}
}