package handlers

import (
	"feishushow/config"
	"feishushow/demo"
	"feishushow/feishu"
	"feishushow/models"
	"feishushow/skills"
	"feishushow/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type SkillHandler struct {
	DB  *gorm.DB
	Cfg *config.Config
}

func (h *SkillHandler) Minutes(c *gin.Context) {
	user, ok := loadUser(c, h.DB)
	if !ok {
		return
	}
	var req struct {
		EventID    string `json:"eventId"`
		CalendarID string `json:"calendarId"`
		WriteBack  bool   `json:"writeBack"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.EventID == "" {
		utils.BadRequest(c, "请选择一场会议")
		return
	}
	now := nowInShanghai()
	var ev *feishu.Event
	if user.Demo {
		ev = demo.EventByID(req.EventID, now)
	} else {
		if err := ensureToken(h.Cfg, h.DB, user); err != nil {
			utils.Unauthorized(c, err.Error())
			return
		}
		cal := req.CalendarID
		if cal == "" {
			cal = "primary"
		}
		got, err := feishu.NewClient(user.AccessToken).GetEvent(cal, req.EventID, now)
		if err != nil {
			utils.BadRequest(c, "读取日程失败："+err.Error())
			return
		}
		ev = got
	}
	if ev == nil {
		utils.BadRequest(c, "找不到这场会议")
		return
	}

	result := skills.BuildMinutes(h.Cfg, *ev, now)

	if req.WriteBack && !user.Demo {
		client := feishu.NewClient(user.AccessToken)
		if url, err := client.CreateDocument(result.Title, result.Markdown); err == nil {
			result.DocURL = url
			result.Source = result.Source + "+docx"
		}
		for _, a := range result.Actions {
			_ = client.CreateTask(a.Title)
		}
	}

	job := models.MinuteJob{
		UserID:   user.ID,
		EventID:  ev.ID,
		Title:    result.Title,
		Markdown: result.Markdown,
		DocURL:   result.DocURL,
	}
	_ = h.DB.Create(&job).Error
	utils.OK(c, result)
}

func (h *SkillHandler) History(c *gin.Context) {
	user, ok := loadUser(c, h.DB)
	if !ok {
		return
	}
	var jobs []models.MinuteJob
	h.DB.Where("user_id = ?", user.ID).Order("id desc").Limit(20).Find(&jobs)
	utils.OK(c, jobs)
}
