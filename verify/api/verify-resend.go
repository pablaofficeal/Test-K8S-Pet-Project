package api

import (
	"Test-K8s-Pet-Project-Verify/providers"
	"Test-K8s-Pet-Project-Verify/configs"
	"Test-K8s-Pet-Project-Verify/core"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
)

func VerifyResend(gdb *gorm.DB, provider providers.Provider, cfg configs.VerifyConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := core.Issue(c.Request.Context(), gdb, provider, cfg, c.GetString("userUUID")); err != nil {
			RespondError(c, err)
			return
		}
		c.JSON(http.StatusAccepted, gin.H{"resend_in": int(cfg.ResendCooldown.Seconds())})
	}
}