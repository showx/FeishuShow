package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func OK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": data, "message": "ok"})
}

func Fail(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{"code": status, "data": nil, "message": message})
}

func BadRequest(c *gin.Context, message string) {
	Fail(c, http.StatusBadRequest, message)
}

func Unauthorized(c *gin.Context, message string) {
	Fail(c, http.StatusUnauthorized, message)
}

func ServerError(c *gin.Context, message string) {
	Fail(c, http.StatusInternalServerError, message)
}
