package cache

import "context"

func DeletePosition(ctx context.Context, boardID string, nodeID string) error {
	pipe := Client.TxPipeline()
	pipe.HDel(ctx, "todo:pos:"+boardID, nodeID)
	pipe.Del(ctx, "todo:node:"+nodeID)
	_, err := pipe.Exec(ctx)
	return err
}