package middlewaretest

import (
	"Test-K8s-Pet-Project-Backend/utils/middleware"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestTraceMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cases := []struct {
		name     string
		header   string
		wantSame bool
	}{
		{"valid id is kept", "req-123_A.b", true},
		{"max length is kept", strings.Repeat("a", 128), true},
		{"empty gets generated", "", false},
		{"too long is replaced", strings.Repeat("a", 129), false},
		{"space is replaced", "bad id", false},
		{"newline is replaced", "bad\nid", false},
		{"slash is replaced", "a/b", false},
		{"unicode is replaced", "идентификатор", false},
	}
	for _, tc := range cases {
		var ctxID any
		r := gin.New()
		r.Use(middleware.TraceMiddleware())
		r.GET("/t", func(c *gin.Context) {
			ctxID, _ = c.Get("trace_id")
			c.Status(200)
		})
		req := httptest.NewRequest("GET", "/t", nil)
		if tc.header != "" {
			req.Header.Set("X-Request-ID", tc.header)
		}
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		got := w.Header().Get("X-Request-ID")
		if got == "" || ctxID != got {
			t.Errorf("%s: header %q and context %v must match and be non-empty", tc.name, got, ctxID)
		}
		if tc.wantSame && got != tc.header {
			t.Errorf("%s: id was changed to %q", tc.name, got)
		}
		if !tc.wantSame {
			if _, err := uuid.Parse(got); err != nil {
				t.Errorf("%s: expected generated uuid, got %q", tc.name, got)
			}
		}
	}
}