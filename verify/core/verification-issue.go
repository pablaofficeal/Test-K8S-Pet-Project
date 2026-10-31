package core

import (
	"Test-K8s-Pet-Project-Verify/providers"
	"Test-K8s-Pet-Project-Verify/configs"
	"Test-K8s-Pet-Project-Verify/models"
	"gorm.io/gorm"
	"context"
	"errors"
	"time"
	"fmt"
)

func Issue(ctx context.Context, gdb *gorm.DB, provider providers.Provider, cfg configs.VerifyConfig, userUUID string) error {
	tx := gdb.WithContext(ctx)

	var user models.User
	if err := tx.First(&user, "uuid = ?", userUUID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUserNotFound
		}
		return err
	}
	if user.EmailVerified {
		return ErrAlreadyVerified
	}

	now := time.Now()
	var last models.EmailVerification
	err := tx.Where("user_uuid = ?", userUUID).Order("created_at DESC").First(&last).Error
	if err == nil {
		if wait := last.CreatedAt.Add(cfg.ResendCooldown).Sub(now); wait > 0 {
			return &CooldownError{RetryAfter: wait}
		}
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	var sent int64
	if err := tx.Model(&models.EmailVerification{}).Where("user_uuid = ? AND created_at > ?", userUUID, now.Add(-time.Hour)).Count(&sent).Error; err != nil {
		return err
	}
	if int(sent) >= cfg.MaxSendsPerHour {
		return ErrRateLimited
	}

	code, err := GenerateCode(cfg)
	if err != nil {
		return err
	}

	verification := models.EmailVerification{
		UserUUID:  userUUID,
		Email:     user.Email,
		CodeHash:  HashCode(cfg.Pepper, userUUID, code),
		ExpiresAt: now.Add(cfg.CodeTTL),
	}
	err = tx.Transaction(func(inner *gorm.DB) error {
		if err := inner.Model(&models.EmailVerification{}).Where("user_uuid = ? AND consumed_at IS NULL", userUUID).Update("consumed_at", now).Error; err != nil {
			return err
		}
		return inner.Create(&verification).Error
	})
	if err != nil {
		return err
	}

	msg := providers.Message{To: user.Email, Username: user.Username, Code: code, ExpiresIn: cfg.CodeTTL}
	if err := provider.Send(ctx, msg); err != nil {
		tx.Delete(&verification)
		return fmt.Errorf("%w: %v", ErrSendFailed, err)
	}
	return nil
}