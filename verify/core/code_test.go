package core

import (
	"Test-K8s-Pet-Project-Verify/configs"
	"testing"
	"regexp"
)

func TestGenerateCodeTestProviderIsFixed(t *testing.T) {
	cfg := configs.VerifyConfig{Provider: "test", TestCode: "893658"}
	for i := 0; i < 5; i++ {
		code, err := GenerateCode(cfg)
		if err != nil || code != "893658" {
			t.Fatalf("expected the fixed test code, got %q (%v)", code, err)
		}
	}
}

func TestGenerateCodeRealProviderIsRandomSixDigits(t *testing.T) {
	cfg := configs.VerifyConfig{Provider: "smtp", TestCode: "893658"}
	format := regexp.MustCompile("^[0-9]{6}$")
	seen := map[string]bool{}
	for i := 0; i < 200; i++ {
		code, err := GenerateCode(cfg)
		if err != nil || !format.MatchString(code) {
			t.Fatalf("expected 6 digits, got %q (%v)", code, err)
		}
		seen[code] = true
	}
	if len(seen) < 150 {
		t.Fatalf("codes are not random enough: %d unique of 200", len(seen))
	}
}

func TestHashCodeBindsUserAndPepper(t *testing.T) {
	base := HashCode("pepper-1", "user-a", "893658")
	if base != HashCode("pepper-1", "user-a", "893658") {
		t.Fatal("hash must be deterministic")
	}
	if base == HashCode("pepper-1", "user-b", "893658") {
		t.Fatal("hash must differ per user")
	}
	if base == HashCode("pepper-2", "user-a", "893658") {
		t.Fatal("hash must differ per pepper")
	}
	if base == HashCode("pepper-1", "user-a", "893659") {
		t.Fatal("hash must differ per code")
	}
}