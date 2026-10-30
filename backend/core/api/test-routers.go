package api

import (
	"github.com/gin-gonic/gin"
)

func TestRouters(c *gin.Context) {
	c.JSON(200, gin.H{"message": "API is working"})
}