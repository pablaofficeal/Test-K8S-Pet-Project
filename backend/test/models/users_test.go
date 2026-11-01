package modelstest

import (
	"Test-K8s-Pet-Project-Backend/models"
	"gorm.io/gorm/schema"
	"encoding/json"
	"strings"
	"testing"
	"sync"
	"time"
)

func TestUserJSONHidesPassword(t *testing.T) {
	now := time.Now()
	raw, err := json.Marshal(models.User{UUID: "u1", Username: "n", Email: "e@x.com", Password: "secret-hash", EmailVerifiedAt: &now})
	if err != nil {
		t.Fatal(err)
	}
	s := string(raw)
	if strings.Contains(s, "secret-hash") || strings.Contains(strings.ToLower(s), "password") {
		t.Errorf("password leaked into JSON: %s", s)
	}
	for _, key := range []string{"is_active", "email_verified", "email_verified_at", "created_at", "updated_at"} {
		if !strings.Contains(s, `"`+key+`"`) {
			t.Errorf("missing json key %q in %s", key, s)
		}
	}
}

func TestUserTableName(t *testing.T) {
	s, err := schema.Parse(&models.User{}, &sync.Map{}, schema.NamingStrategy{})
	if err != nil {
		t.Fatal(err)
	}
	if s.Table != "users" {
		t.Errorf("table = %q, want users", s.Table)
	}
}