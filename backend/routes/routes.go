package routes

import (
	"feishushow/config"
	"feishushow/handlers"
	"feishushow/middleware"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func Setup(r *gin.Engine, db *gorm.DB, cfg *config.Config) {
	r.Use(middleware.CORS(cfg))

	auth := &handlers.AuthHandler{DB: db, Cfg: cfg}
	today := &handlers.TodayHandler{DB: db, Cfg: cfg}
	skill := &handlers.SkillHandler{DB: db, Cfg: cfg}

	api := r.Group("/api")
	{
		api.GET("/auth/status", auth.Status)
		api.GET("/auth/feishu/url", auth.FeishuURL)
		api.POST("/auth/feishu/exchange", auth.FeishuExchange)
		api.POST("/auth/demo", auth.Demo)

		need := api.Group("")
		need.Use(middleware.Auth(cfg))
		{
			need.GET("/me", auth.Me)
			need.GET("/today", today.Get)
			need.POST("/skills/minutes", skill.Minutes)
			need.GET("/skills/minutes", skill.History)
		}
	}
}
