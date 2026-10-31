package core

import (
	"Test-K8s-Pet-Project-Verify/configs"
	"Test-K8s-Pet-Project-Verify/models"
	"crypto/subtle"
	"gorm.io/gorm"
	"context"
	"strings"
	"errors"
	"time"
)

func Confirm(ctx context.Context, gdb *gorm.DB, cfg configs.VerifyConfig, userUUID string, code string) error {
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
	var verification models.EmailVerification
	err := tx.Where("user_uuid = ? AND consumed_at IS NULL AND expires_at > ?", userUUID, now).Order("created_at DESC").First(&verification).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNoActiveCode
		}
		return err
	}
	if !strings.EqualFold(verification.Email, user.Email) {
		return ErrNoActiveCode
	}

	// Попытка списывается атомарно до сравнения, иначе параллельные запросы обходят лимит
	spent := tx.Model(&models.EmailVerification{}).Where("id = ? AND attempts < ?", verification.ID, cfg.MaxAttempts).UpdateColumn("attempts", gorm.Expr("attempts + 1"))
	if spent.Error != nil {
		return spent.Error
	}
	if spent.RowsAffected == 0 {
		return ErrTooManyAttempts
	}

	expected := HashCode(cfg.Pepper, userUUID, code)
	if subtle.ConstantTimeCompare([]byte(expected), []byte(verification.CodeHash)) != 1 {
		return &InvalidCodeError{AttemptsLeft: max(cfg.MaxAttempts-verification.Attempts-1, 0)}
	}

	return tx.Transaction(func(inner *gorm.DB) error {
		if err := inner.Model(&models.EmailVerification{}).Where("id = ?", verification.ID).Update("consumed_at", now).Error; err != nil {
			return err
		}
		return inner.Model(&models.User{}).Where("uuid = ?", userUUID).Updates(map[string]any{"email_verified": true, "email_verified_at": now}).Error
	})
}