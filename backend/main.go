package main

import (
	"log"
	"os"

	"feishushow/config"
	"feishushow/database"
	"feishushow/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()
	db, err := database.Init(cfg.DBPath)
	if err != nil {
		log.Fatalf("数据库初始化失败: %v", err)
	}

	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.Default()
	r.SetTrustedProxies(nil)
	routes.Setup(r, db, cfg)

	addr := ":" + cfg.Port
	log.Printf("FeishuShow API 已启动 http://localhost%s", addr)
	if cfg.FeishuConfigured() {
		log.Printf("飞书登录已启用，回调 %s", cfg.FeishuRedirect)
	} else {
		log.Printf("未配置飞书应用，可先用演示模式")
	}
	if err := r.Run(addr); err != nil {
		log.Fatal(err)
	}
}
