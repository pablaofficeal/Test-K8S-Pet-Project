package api

import (
	"Test-K8s-Pet-Project-Verify/configs"
	"Test-K8s-Pet-Project-Verify/core"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"strings"
	"regexp"
	"errors"
)

type confirmRequest struct {
	Code string `json:"code"`
}

var codeFormat = regexp.MustCompile("^[0-9]{6}$")

func VerifyConfirm(gdb *gorm.DB, cfg configs.VerifyConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		var body confirmRequest
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}

		code := strings.NewReplacer(" ", "", "-", "").Replace(body.Code)
		if !codeFormat.MatchString(code) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "code must be 6 digits"})
			return
		}

		err := core.Confirm(c.Request.Context(), gdb, cfg, c.GetString("userUUID"), code)
		if err == nil || errors.Is(err, core.ErrAlreadyVerified) {
			c.JSON(http.StatusOK, gin.H{"verified": true})
			return
		}
		RespondError(c, err)
	}
}