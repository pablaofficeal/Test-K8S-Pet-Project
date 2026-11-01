package models

import "time"

type MessengerMessage struct {
	ID int64 `json:"id" gorm:"primaryKey;autoIncrement"`
	SenderUUID string `json:"sender_uuid" gorm:"type:uuid;not null;index:idx_messenger_pair,priority:1"`
	RecipientUUID string `json:"recipient_uuid" gorm:"type:uuid;not null;index:idx_messenger_pair,priority:2;index:idx_messenger_inbox,priority:1"`
	Envelope string `json:"envelope" gorm:"type:jsonb;not null"`
	ReadAt *time.Time `json:"read_at" gorm:"index:idx_messenger_inbox,priority:2"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
}
