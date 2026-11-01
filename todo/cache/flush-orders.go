package cache

import (
	"Cadence-Cloud-Todo/models"
	"Cadence-Cloud-Todo/db"
	"gorm.io/gorm"
	"context"
	"log"
)

func FlushOrders(ctx context.Context) (int, error) {
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
			ids, found, err := GetOrder(ctx, boardID)
			if err == nil && found {
				err = db.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
					for index, nodeID := range ids {
						if err := tx.Model(&models.TodoNode{}).Where("id = ? AND board_id = ?", nodeID, boardID).Update("sort_order", index).Error; err != nil {
							return err
						}
					}
					return nil
				})
			}
			if err != nil {
				log.Printf("flush order of board %s failed: %v", boardID, err)
				Client.SAdd(ctx, "todo:dirty", boardID)
				return flushed, err
			}
			flushed += len(ids)
		}
	}
}