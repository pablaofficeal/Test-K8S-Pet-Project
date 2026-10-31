package configs

import (
	"testing"
	"os"
)

func setEnv(t *testing.T, values map[string]string) {
	for _, key := range []string{"VERIFY_PROVIDER", "VERIFY_TEST_CODE", "VERIFY_ALLOW_TEST_CODE", "VERIFY_CODE_PEPPER", "VERIFY_CODE_TTL", "VERIFY_MAX_ATTEMPTS"} {
		t.Setenv(key, "")
		os.Unsetenv(key)
	}
	for key, value := range values {
		t.Setenv(key, value)
	}
}

func TestLoadVerifyConfigDefaults(t *testing.T) {
	setEnv(t, map[string]string{"VERIFY_CODE_PEPPER": "long-enough-pepper", "VERIFY_ALLOW_TEST_CODE": "true"})
	cfg, err := LoadVerifyConfig()
	if err != nil || cfg.TestCode != "893658" || cfg.MaxAttempts != 5 || cfg.CodeTTL.Minutes() != 10 {
		t.Fatalf("unexpected defaults: %+v (%v)", cfg, err)
	}
}

func TestLoadVerifyConfigRefusesTestProviderWithoutExplicitAllow(t *testing.T) {
	setEnv(t, map[string]string{"VERIFY_CODE_PEPPER": "long-enough-pepper"})
	if _, err := LoadVerifyConfig(); err == nil {
		t.Fatal("the fixed test code must be opt-in")
	}
}

func TestLoadVerifyConfigValidatesValues(t *testing.T) {
	cases := map[string]map[string]string{
		"short pepper":  {"VERIFY_CODE_PEPPER": "short", "VERIFY_ALLOW_TEST_CODE": "true"},
		"bad test code": {"VERIFY_CODE_PEPPER": "long-enough-pepper", "VERIFY_ALLOW_TEST_CODE": "true", "VERIFY_TEST_CODE": "12345"},
		"bad ttl":       {"VERIFY_CODE_PEPPER": "long-enough-pepper", "VERIFY_ALLOW_TEST_CODE": "true", "VERIFY_CODE_TTL": "soon"},
		"bad attempts":  {"VERIFY_CODE_PEPPER": "long-enough-pepper", "VERIFY_ALLOW_TEST_CODE": "true", "VERIFY_MAX_ATTEMPTS": "0"},
	}
	for name, values := range cases {
		setEnv(t, values)
		if _, err := LoadVerifyConfig(); err == nil {
			t.Errorf("%s: expected an error", name)
		}
	}
}