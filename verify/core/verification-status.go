package core

import (
	"Test-K8s-Pet-Project-Verify/configs"
	"Test-K8s-Pet-Project-Verify/models"
	"gorm.io/gorm"
	"context"
	"strings"
	"errors"
	"math"
	"time"
)

type StatusResult struct {
	Verified  bool
	Email     string
	ResendIn  int
	ExpiresIn int
}

func Status(ctx context.Context, gdb *gorm.DB, cfg configs.VerifyConfig, userUUID string) (StatusResult, error) {
	tx := gdb.WithContext(ctx)

	var user models.User
	if err := tx.First(&user, "uuid = ?", userUUID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return StatusResult{}, ErrUserNotFound
		}
		return StatusResult{}, err
	}

	result := StatusResult{Verified: user.EmailVerified, Email: user.Email}
	if user.EmailVerified {
		return result, nil
	}

	now := time.Now()
	var last models.EmailVerification
	err := tx.Where("user_uuid = ?", userUUID).Order("created_at DESC").First(&last).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return result, nil
	}
	if err != nil {
		return result, err
	}

	if wait := last.CreatedAt.Add(cfg.ResendCooldown).Sub(now); wait > 0 {
		result.ResendIn = int(math.Ceil(wait.Seconds()))
	}
	if last.ConsumedAt == nil && last.ExpiresAt.After(now) && strings.EqualFold(last.Email, user.Email) {
		result.ExpiresIn = int(math.Ceil(last.ExpiresAt.Sub(now).Seconds()))
	}
	return result, nil
}