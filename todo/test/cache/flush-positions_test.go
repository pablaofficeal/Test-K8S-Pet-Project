package cachetest

import (
	"Cadence-Cloud-Todo/models"
	"Cadence-Cloud-Todo/cache"
	"github.com/google/uuid"
	"Cadence-Cloud-Todo/db"
	"context"
	"testing"
	"time"
)

func TestFlushWithoutDirtyBoards(t *testing.T) {
	redisServer.FlushAll()
	n, err := cache.FlushPositions(context.Background())
	if err != nil || n != 0 {
		t.Errorf("flush = %d, %v", n, err)
	}
}

func TestFlushPositionsWritesToPostgres(t *testing.T) {
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
	a := models.TodoNode{BoardID: board.ID, Title: "a"}
	b := models.TodoNode{BoardID: board.ID, Title: "b"}
	tx.Create(&a)
	tx.Create(&b)

	cache.SetPosition(ctx, board.ID, a.ID, 111, 222)
	cache.SetPosition(ctx, board.ID, a.ID, 150, 250)

	var before models.TodoNode
	tx.First(&before, "id = ?", a.ID)
	if before.PosX != 0 || before.PosY != 0 {
		t.Fatalf("position reached postgres before the flush: %v,%v", before.PosX, before.PosY)
	}

	n, err := cache.FlushPositions(ctx)
	if err != nil || n != 1 {
		t.Fatalf("flush = %d, %v", n, err)
	}
	var after, untouched models.TodoNode
	tx.First(&after, "id = ?", a.ID)
	tx.First(&untouched, "id = ?", b.ID)
	if after.PosX != 150 || after.PosY != 250 {
		t.Errorf("flushed position = %v,%v, want the last write 150,250", after.PosX, after.PosY)
	}
	if untouched.PosX != 0 || untouched.PosY != 0 {
		t.Errorf("a node without a hot position was modified: %v,%v", untouched.PosX, untouched.PosY)
	}
	if members, _ := redisServer.SMembers("todo:dirty"); len(members) != 0 {
		t.Errorf("dirty set must be empty after the flush, got %v", members)
	}
	if got, _ := cache.GetPositions(ctx, board.ID); got[a.ID] != [2]float64{150, 250} {
		t.Error("hot position must stay readable after the flush")
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
	foreign := models.TodoNode{BoardID: theirs.ID, Title: "foreign"}
	tx.Create(&foreign)

	cache.SetPosition(ctx, mine.ID, foreign.ID, 5, 5)
	cache.FlushPositions(ctx)

	var got models.TodoNode
	tx.First(&got, "id = ?", foreign.ID)
	if got.PosX != 0 || got.PosY != 0 {
		t.Error("flush wrote a position into a node that belongs to another board")
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
	node := models.TodoNode{BoardID: board.ID, Title: "n"}
	tx.Create(&node)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	cache.StartFlusher(ctx, 30*time.Millisecond, done)
	cache.SetPosition(context.Background(), board.ID, node.ID, 7, 8)
	time.Sleep(300 * time.Millisecond)

	var got models.TodoNode
	tx.First(&got, "id = ?", node.ID)
	if got.PosX != 7 || got.PosY != 8 {
		t.Errorf("periodic flush did not happen: %v,%v", got.PosX, got.PosY)
	}

	cache.SetPosition(context.Background(), board.ID, node.ID, 70, 80)
	cancel()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("flusher did not stop")
	}
	tx.First(&got, "id = ?", node.ID)
	if got.PosX != 70 || got.PosY != 80 {
		t.Errorf("final flush on shutdown did not happen: %v,%v", got.PosX, got.PosY)
	}
}