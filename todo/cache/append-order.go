package cache

import "context"

func AppendToOrder(ctx context.Context, boardID string, nodeID string) error {
	ids, found, err := GetOrder(ctx, boardID)
	if err != nil || !found {
		return err
	}
	return SetOrder(ctx, boardID, append(ids, nodeID))
}