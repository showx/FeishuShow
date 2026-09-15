package demo

import (
	"time"

	"feishushow/feishu"
)

func Board(now time.Time) feishu.TodayBoard {
	loc := now.Location()
	at := func(h, m int) time.Time {
		return time.Date(now.Year(), now.Month(), now.Day(), h, m, 0, 0, loc)
	}

	events := []feishu.Event{
		{
			ID: "demo-standup", CalendarID: "demo", Title: "产品站会 · FeishuShow",
			Start: at(9, 30), End: at(9, 50), Location: "飞书会议",
			Description: "同步昨日进度和今日风险。",
			Attendees:   []string{"陈晨", "林可", "王路"},
			HasMeeting:  true, Status: feishu.ClassifyStatus(at(9, 30), at(9, 50), now),
		},
		{
			ID: "demo-review", CalendarID: "demo", Title: "竞品复盘：豆包工作 vs 开源雷达",
			Start: at(11, 0), End: at(12, 0), Location: "3F 会议室 / 飞书会议",
			Description: "对齐差异化：只做「看清飞书」，不抢官方 Agent。",
			Attendees:   []string{"陈晨", "周岩", "苏青"},
			HasMeeting:  true, Status: feishu.ClassifyStatus(at(11, 0), at(12, 0), now),
		},
		{
			ID: "demo-customer", CalendarID: "demo", Title: "客户走访准备：星河科技",
			Start: at(15, 0), End: at(16, 0), Location: "飞书文档协作",
			Description: "带上上周会议纪要和报价草稿。",
			Attendees:   []string{"陈晨", "客户成功组"},
			HasMeeting:  false, Status: feishu.ClassifyStatus(at(15, 0), at(16, 0), now),
		},
		{
			ID: "demo-ended", CalendarID: "demo", Title: "周初同步 · 工程周会",
			Start: at(8, 0), End: at(8, 40), Location: "飞书会议",
			Description: "版本节奏、阻塞项、下周里程碑。",
			Attendees: []string{"陈晨", "后端", "前端"},
			HasMeeting: true, Status: feishu.ClassifyStatus(at(8, 0), at(8, 40), now),
		},
	}
	if events[3].Status != "ended" {
		events[3].Start = now.Add(-90 * time.Minute)
		events[3].End = now.Add(-30 * time.Minute)
		events[3].Status = "ended"
	}

	dueSoon := now.Add(6 * time.Hour)
	overdue := now.Add(-26 * time.Hour)
	tasks := []feishu.Task{
		{ID: "t1", Title: "补齐飞书应用权限清单到 README", Due: &dueSoon},
		{ID: "t2", Title: "把会后纪要技能接到豆包 API", Due: &dueSoon},
		{ID: "t3", Title: "跟进星河科技报价口径", Due: &overdue, Overdue: true},
		{ID: "t4", Title: "文档雷达：标出发呆 30 天以上的知识库"},
	}

	docs := []feishu.Doc{
		{Token: "d1", Name: "FeishuShow 产品一页纸", Type: "docx", ModifiedAt: now.Add(-2 * time.Hour), DaysIdle: 0},
		{Token: "d2", Name: "2026 Q3 客户拜访记录", Type: "docx", ModifiedAt: now.Add(-40 * 24 * time.Hour), Stale: true, DaysIdle: 40},
		{Token: "d3", Name: "豆包工作能力对照表", Type: "sheet", ModifiedAt: now.Add(-5 * 24 * time.Hour), DaysIdle: 5},
		{Token: "d4", Name: "知识库 / 过期制度汇编", Type: "docx", ModifiedAt: now.Add(-92 * 24 * time.Hour), Stale: true, DaysIdle: 92},
		{Token: "d5", Name: "工程周会纪要模板", Type: "docx", ModifiedAt: now.Add(-12 * time.Hour), DaysIdle: 0},
	}

	chats := []feishu.Chat{
		{ID: "c1", Name: "FeishuShow 核心组", MemberCount: 6, UpdatedAt: now.Add(-20 * time.Minute)},
		{ID: "c2", Name: "客户成功 · 星河科技", MemberCount: 8, UpdatedAt: now.Add(-2 * time.Hour)},
		{ID: "c3", Name: "开源发布通知", MemberCount: 24, UpdatedAt: now.Add(-26 * time.Hour)},
	}

	return feishu.TodayBoard{
		Date:     now.Format("2006-01-02"),
		Stats:    feishu.Summarize(events, tasks, docs, chats),
		Events:   events,
		Tasks:    tasks,
		Docs:     docs,
		Chats:    chats,
		Warnings: []string{"当前是本地演示数据，连接飞书后会替换为你本人权限内的真实日程、文档和任务。"},
	}
}

func EventByID(id string, now time.Time) *feishu.Event {
	board := Board(now)
	for i := range board.Events {
		if board.Events[i].ID == id {
			ev := board.Events[i]
			return &ev
		}
	}
	return nil
}
