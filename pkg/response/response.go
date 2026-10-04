package response

import "github.com/gin-gonic/gin"

func Success(c *gin.Context, code int, message string, data interface{}) {
	c.JSON(code, gin.H{
		"success": true,
		"message": message,
		"data":    data,
	})
}

func Error(c *gin.Context, code int, message string, err error) {
	errStr := ""
	if err != nil {
		errStr = err.Error()
	}
	c.JSON(code, gin.H{
		"success": false,
		"error":   message,
		"details": errStr,
	})
}
