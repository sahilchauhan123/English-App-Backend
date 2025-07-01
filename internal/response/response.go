package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, gin.H{
		"data": data,
	})
}

func Failed(c *gin.Context, status int, reason string) {
	c.JSON(status, gin.H{
		"success": false,
		"error":   reason,
	})
}
