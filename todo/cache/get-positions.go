package cache

import (
	"strconv"
	"strings"
	"context"
)

func GetPositions(ctx context.Context, boardID string) (map[string][2]float64, error) {
	raw, err := Client.HGetAll(ctx, "todo:pos:"+boardID).Result()
	if err != nil {
		return nil, err
	}
	out := make(map[string][2]float64, len(raw))
	for nodeID, value := range raw {
		xs, ys, found := strings.Cut(value, ":")
		if !found {
			continue
		}
		x, errX := strconv.ParseFloat(xs, 64)
		y, errY := strconv.ParseFloat(ys, 64)
		if errX == nil && errY == nil {
			out[nodeID] = [2]float64{x, y}
		}
	}
	return out, nil
}