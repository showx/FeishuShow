package handlers

import (
	"feishushow/config"
	"feishushow/feishu"
	"feishushow/models"
	"feishushow/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AuthHandler struct {
	DB  *gorm.DB
	Cfg *config.Config
}

func (h *AuthHandler) Status(c *gin.Context) {
	utils.OK(c, gin.H{
		"feishuConfigured": h.Cfg.FeishuConfigured(),
		"llmConfigured":    h.Cfg.LLMConfigured(),
		"demoAvailable":    true,
	})
}

func (h *AuthHandler) FeishuURL(c *gin.Context) {
	if !h.Cfg.FeishuConfigured() {
		utils.BadRequest(c, "尚未配置 FEISHU_APP_ID / FEISHU_APP_SECRET，可先用演示模式")
		return
	}
	state := feishu.NewState()
	utils.OK(c, gin.H{"url": feishu.AuthorizeLink(h.Cfg.FeishuAppID, h.Cfg.FeishuRedirect, state), "state": state})
}

func (h *AuthHandler) FeishuExchange(c *gin.Context) {
	if !h.Cfg.FeishuConfigured() {
		utils.BadRequest(c, "尚未配置飞书应用")
		return
	}
	var req struct {
		Code  string `json:"code"`
		State string `json:"state"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Code == "" {
		utils.BadRequest(c, "缺少授权码")
		return
	}
	if !feishu.ConsumeState(req.State) {
		utils.BadRequest(c, "授权状态无效或已过期，请重新登录")
		return
	}
	set, err := feishu.ExchangeCode(h.Cfg.FeishuAppID, h.Cfg.FeishuAppSecret, req.Code, h.Cfg.FeishuRedirect)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	info, err := feishu.FetchUserInfo(set.AccessToken)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	if info.OpenID == "" {
		utils.BadRequest(c, "未拿到用户 open_id")
		return
	}
	user := models.User{OpenID: info.OpenID}
	_ = h.DB.Where("open_id = ?", info.OpenID).First(&user).Error
	user.OpenID = info.OpenID
	user.Name = info.Name
	user.Avatar = info.Avatar
	user.Email = info.Email
	user.TenantKey = info.TenantKey
	user.Demo = false
	applyTokens(&user, set)
	if err := h.DB.Save(&user).Error; err != nil {
		utils.ServerError(c, "保存用户失败")
		return
	}
	issueSession(c, h.Cfg, &user)
}

func (h *AuthHandler) Demo(c *gin.Context) {
	user := models.User{OpenID: "demo-local"}
	err := h.DB.Where("open_id = ?", user.OpenID).First(&user).Error
	if err != nil {
		user = models.User{
			OpenID: "demo-local",
			Name:   "演示同事",
			Avatar: "",
			Email:  "demo@feishushow.local",
			Demo:   true,
		}
		if err := h.DB.Create(&user).Error; err != nil {
			utils.ServerError(c, "创建演示账号失败")
			return
		}
	}
	user.Demo = true
	user.Name = "演示同事"
	_ = h.DB.Save(&user).Error
	issueSession(c, h.Cfg, &user)
}

func (h *AuthHandler) Me(c *gin.Context) {
	user, ok := loadUser(c, h.DB)
	if !ok {
		return
	}
	utils.OK(c, user.Public())
}
