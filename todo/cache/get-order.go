package cache

import (
	"github.com/redis/go-redis/v9"
	"encoding/json"
	"context"
)

func GetOrder(ctx context.Context, boardID string) ([]string, bool, error) {
	raw, err := Client.Get(ctx, "todo:order:"+boardID).Bytes()
	if err == redis.Nil {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	var ids []string
	if err := json.Unmarshal(raw, &ids); err != nil {
		return nil, false, nil
	}
	return ids, true, nil
}