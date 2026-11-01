package cache

import "context"

func RemoveFromOrder(ctx context.Context, boardID string, nodeID string) error {
	ids, found, err := GetOrder(ctx, boardID)
	if err != nil || !found {
		return err
	}
	kept := make([]string, 0, len(ids))
	for _, id := range ids {
		if id != nodeID {
			kept = append(kept, id)
		}
	}
	return SetOrder(ctx, boardID, kept)
}