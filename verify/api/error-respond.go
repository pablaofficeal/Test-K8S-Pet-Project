package api

import (
	"Test-K8s-Pet-Project-Verify/core"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"errors"
	"math"
	"log"
)

func RespondError(c *gin.Context, err error) {
	traceID, _ := c.Get("trace_id")
	var cooldown *core.CooldownError
	var invalid *core.InvalidCodeError

	switch {
	case errors.Is(err, core.ErrUserNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
	case errors.Is(err, core.ErrAlreadyVerified):
		c.JSON(http.StatusConflict, gin.H{"error": "email already verified"})
	case errors.As(err, &cooldown):
		seconds := int(math.Ceil(cooldown.RetryAfter.Seconds()))
		c.Header("Retry-After", strconv.Itoa(seconds))
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "please wait before requesting a new code", "retry_after": seconds})
	case errors.Is(err, core.ErrRateLimited):
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "too many verification requests, try again later"})
	case errors.Is(err, core.ErrNoActiveCode):
		c.JSON(http.StatusBadRequest, gin.H{"error": "no active code, request a new one"})
	case errors.Is(err, core.ErrTooManyAttempts):
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "too many attempts, request a new code"})
	case errors.As(err, &invalid):
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid code", "attempts_left": invalid.AttemptsLeft})
	case errors.Is(err, core.ErrSendFailed):
		log.Printf("trace_id=%s %v", traceID, err)
		c.JSON(http.StatusBadGateway, gin.H{"error": "could not send the code, try again later"})
	default:
		log.Printf("trace_id=%s internal error: %v", traceID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
	}
}