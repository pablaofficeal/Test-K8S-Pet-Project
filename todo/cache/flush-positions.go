package cache

import (
	"Cadence-Cloud-Todo/models"
	"Cadence-Cloud-Todo/db"
	"gorm.io/gorm"
	"context"
	"log"
)

func FlushPositions(ctx context.Context) (int, error) {
	flushed := 0
	for {
		boards, err := Client.SPopN(ctx, "todo:dirty", 100).Result()
		if err != nil {
			return flushed, err
		}
		if len(boards) == 0 {
			return flushed, nil
		}
		for _, boardID := range boards {
			positions, err := GetPositions(ctx, boardID)
			if err == nil {
				err = db.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
					for nodeID, p := range positions {
						update := map[string]any{"pos_x": p[0], "pos_y": p[1]}
						if err := tx.Model(&models.TodoNode{}).Where("id = ? AND board_id = ?", nodeID, boardID).Updates(update).Error; err != nil {
							return err
						}
					}
					return nil
				})
			}
			if err != nil {
				log.Printf("flush positions of board %s failed: %v", boardID, err)
				Client.SAdd(ctx, "todo:dirty", boardID)
				return flushed, err
			}
			flushed += len(positions)
		}
	}
}