package cache

import (
	"strconv"
	"context"
	"time"
)

func SetPosition(ctx context.Context, boardID string, nodeID string, x float64, y float64) error {
	value := strconv.FormatFloat(x, 'f', -1, 64) + ":" + strconv.FormatFloat(y, 'f', -1, 64)
	pipe := Client.TxPipeline()
	pipe.HSet(ctx, "todo:pos:"+boardID, nodeID, value)
	pipe.Expire(ctx, "todo:pos:"+boardID, 24*time.Hour)
	pipe.SAdd(ctx, "todo:dirty", boardID)
	_, err := pipe.Exec(ctx)
	return err
}