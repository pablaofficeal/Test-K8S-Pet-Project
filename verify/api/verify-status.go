package api

import (
	"Test-K8s-Pet-Project-Verify/configs"
	"Test-K8s-Pet-Project-Verify/core"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
)

func VerifyStatus(gdb *gorm.DB, cfg configs.VerifyConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		result, err := core.Status(c.Request.Context(), gdb, cfg, c.GetString("userUUID"))
		if err != nil {
			RespondError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"verified":   result.Verified,
			"email":      result.Email,
			"resend_in":  result.ResendIn,
			"expires_in": result.ExpiresIn,
		})
	}
}