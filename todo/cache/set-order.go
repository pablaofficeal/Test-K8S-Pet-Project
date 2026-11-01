package cache

import (
	"encoding/json"
	"context"
	"time"
)

func SetOrder(ctx context.Context, boardID string, nodeIDs []string) error {
	raw, err := json.Marshal(nodeIDs)
	if err != nil {
		return err
	}
	pipe := Client.TxPipeline()
	pipe.Set(ctx, "todo:order:"+boardID, raw, 24*time.Hour)
	pipe.SAdd(ctx, "todo:dirty", boardID)
	_, err = pipe.Exec(ctx)
	return err
}