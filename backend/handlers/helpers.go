package handlers

import (
	"time"

	"feishushow/config"
	"feishushow/feishu"
	"feishushow/middleware"
	"feishushow/models"
	"feishushow/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func nowInShanghai() time.Time {
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		return time.Now()
	}
	return time.Now().In(loc)
}

func loadUser(c *gin.Context, db *gorm.DB) (*models.User, bool) {
	var user models.User
	if err := db.First(&user, middleware.CurrentUserID(c)).Error; err != nil {
		utils.Unauthorized(c, "用户不存在，请重新登录")
		return nil, false
	}
	return &user, true
}

func ensureToken(cfg *config.Config, db *gorm.DB, user *models.User) error {
	if user.Demo || user.AccessToken == "" {
		return nil
	}
	if time.Until(user.TokenExpiry) > 3*time.Minute {
		return nil
	}
	if user.RefreshToken == "" || time.Now().After(user.RefreshExpiry) {
		return errNeedReauth
	}
	set, err := feishu.Refresh(cfg.FeishuAppID, cfg.FeishuAppSecret, user.RefreshToken)
	if err != nil {
		return err
	}
	applyTokens(user, set)
	return db.Save(user).Error
}

var errNeedReauth = errString("飞书登录已过期，请重新授权")

type errString string

func (e errString) Error() string { return string(e) }

func applyTokens(user *models.User, set *feishu.TokenSet) {
	user.AccessToken = set.AccessToken
	if set.RefreshToken != "" {
		user.RefreshToken = set.RefreshToken
	}
	exp := set.ExpiresIn
	if exp <= 0 {
		exp = 7200
	}
	user.TokenExpiry = time.Now().Add(time.Duration(exp) * time.Second)
	if set.RefreshExpiresIn > 0 {
		user.RefreshExpiry = time.Now().Add(time.Duration(set.RefreshExpiresIn) * time.Second)
	} else if user.RefreshExpiry.IsZero() {
		user.RefreshExpiry = time.Now().Add(7 * 24 * time.Hour)
	}
}

func issueSession(c *gin.Context, cfg *config.Config, user *models.User) {
	token, err := utils.SignToken(cfg.JWTSecret, user.ID, user.Name, user.Demo, cfg.JWTExpire)
	if err != nil {
		utils.ServerError(c, "签发登录态失败")
		return
	}
	utils.OK(c, gin.H{"token": token, "user": user.Public()})
}
