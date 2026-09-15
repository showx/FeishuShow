package handlers

import (
	"feishushow/config"
	"feishushow/demo"
	"feishushow/feishu"
	"feishushow/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type TodayHandler struct {
	DB  *gorm.DB
	Cfg *config.Config
}

func (h *TodayHandler) Get(c *gin.Context) {
	user, ok := loadUser(c, h.DB)
	if !ok {
		return
	}
	now := nowInShanghai()
	if user.Demo {
		utils.OK(c, demo.Board(now))
		return
	}
	if err := ensureToken(h.Cfg, h.DB, user); err != nil {
		utils.Unauthorized(c, err.Error())
		return
	}
	board := feishu.NewClient(user.AccessToken).LoadToday(now)
	utils.OK(c, board)
}
