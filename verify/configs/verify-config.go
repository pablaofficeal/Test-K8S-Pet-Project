package configs

import (
	"regexp"
	"errors"
	"time"
)

var testCodePattern = regexp.MustCompile("^[0-9]{6}$")

type VerifyConfig struct {
	Provider        string
	TestCode        string
	AllowTestCode   bool
	Pepper          string
	CodeTTL         time.Duration
	ResendCooldown  time.Duration
	MaxAttempts     int
	MaxSendsPerHour int
}

func LoadVerifyConfig() (VerifyConfig, error) {
	cfg := VerifyConfig{
		Provider:      EnvString("VERIFY_PROVIDER", "test"),
		TestCode:      EnvString("VERIFY_TEST_CODE", "893658"),
		AllowTestCode: EnvString("VERIFY_ALLOW_TEST_CODE", "false") == "true",
		Pepper:        EnvString("VERIFY_CODE_PEPPER", ""),
	}
	var err error
	if cfg.CodeTTL, err = EnvDuration("VERIFY_CODE_TTL", 10*time.Minute); err != nil {
		return cfg, err
	}
	if cfg.ResendCooldown, err = EnvDuration("VERIFY_RESEND_COOLDOWN", 60*time.Second); err != nil {
		return cfg, err
	}
	if cfg.MaxAttempts, err = EnvInt("VERIFY_MAX_ATTEMPTS", 5); err != nil {
		return cfg, err
	}
	if cfg.MaxSendsPerHour, err = EnvInt("VERIFY_MAX_SENDS_PER_HOUR", 5); err != nil {
		return cfg, err
	}
	if len(cfg.Pepper) < 8 {
		return cfg, errors.New("VERIFY_CODE_PEPPER must be at least 8 characters")
	}
	if cfg.Provider == "test" && !cfg.AllowTestCode {
		return cfg, errors.New("the test provider uses a fixed code, set VERIFY_ALLOW_TEST_CODE=true to allow it")
	}
	if cfg.Provider == "test" && !testCodePattern.MatchString(cfg.TestCode) {
		return cfg, errors.New("VERIFY_TEST_CODE must be exactly 6 digits")
	}
	return cfg, nil
}