package middleware

import (
	"strings"

	"feishushow/config"
	"feishushow/utils"

	"github.com/gin-gonic/gin"
)

const ContextUserID = "userId"

func Auth(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			utils.Unauthorized(c, "未登录")
			c.Abort()
			return
		}
		tokenStr := strings.TrimPrefix(header, "Bearer ")
		if tokenStr == header {
			utils.Unauthorized(c, "无效的认证信息")
			c.Abort()
			return
		}
		claims, err := utils.ParseToken(cfg.JWTSecret, tokenStr)
		if err != nil {
			utils.Unauthorized(c, "登录已过期，请重新登录")
			c.Abort()
			return
		}
		c.Set(ContextUserID, claims.UserID)
		c.Next()
	}
}

func CurrentUserID(c *gin.Context) uint {
	v, _ := c.Get(ContextUserID)
	id, _ := v.(uint)
	return id
}
