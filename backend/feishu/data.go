package feishu

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

func (c *Client) LoadToday(now time.Time) TodayBoard {
	loc := now.Location()
	dayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	dayEnd := dayStart.Add(24 * time.Hour)
	board := TodayBoard{
		Date:     now.Format("2006-01-02"),
		Events:   []Event{},
		Tasks:    []Task{},
		Docs:     []Doc{},
		Chats:    []Chat{},
		Warnings: []string{},
	}

	events, err := c.ListTodayEvents(dayStart, dayEnd, now)
	if err != nil {
		board.Warnings = append(board.Warnings, "日程："+friendlyErr(err))
	} else {
		board.Events = events
	}

	tasks, err := c.ListOpenTasks(now)
	if err != nil {
		board.Warnings = append(board.Warnings, "待办："+friendlyErr(err))
	} else {
		board.Tasks = tasks
	}

	docs, err := c.ListRecentDocs(now)
	if err != nil {
		board.Warnings = append(board.Warnings, "文档："+friendlyErr(err))
	} else {
		board.Docs = docs
	}

	chats, err := c.ListRecentChats()
	if err != nil {
		board.Warnings = append(board.Warnings, "会话："+friendlyErr(err))
	} else {
		board.Chats = chats
	}

	board.Stats = Summarize(board.Events, board.Tasks, board.Docs, board.Chats)
	return board
}

func Summarize(events []Event, tasks []Task, docs []Doc, chats []Chat) Stats {
	s := Stats{Meetings: len(events), Tasks: len(tasks), Chats: len(chats)}
	for _, e := range events {
		if e.Status == "live" {
			s.Live++
		}
		if e.Status == "ended" {
			s.Ended++
		}
	}
	for _, t := range tasks {
		if t.Overdue {
			s.Overdue++
		}
	}
	for _, d := range docs {
		if d.Stale {
			s.StaleDocs++
		}
	}
	return s
}

func (c *Client) ListTodayEvents(dayStart, dayEnd, now time.Time) ([]Event, error) {
	var primary struct {
		Calendars []struct {
			Calendar struct {
				CalendarID string `json:"calendar_id"`
			} `json:"calendar"`
		} `json:"calendars"`
	}
	if err := c.post("/calendar/v4/calendars/primary", map[string]any{}, &primary); err != nil {
		return nil, err
	}
	if len(primary.Calendars) == 0 {
		return []Event{}, nil
	}
	calID := primary.Calendars[0].Calendar.CalendarID
	q := url.Values{}
	q.Set("start_time", strconv.FormatInt(dayStart.Unix(), 10))
	q.Set("end_time", strconv.FormatInt(dayEnd.Unix(), 10))
	q.Set("page_size", "50")
	var list struct {
		Items []rawEvent `json:"items"`
	}
	if err := c.get("/calendar/v4/calendars/"+url.PathEscape(calID)+"/events", q, &list); err != nil {
		return nil, err
	}
	out := make([]Event, 0, len(list.Items))
	for _, item := range list.Items {
		if item.Status == "cancelled" {
			continue
		}
		ev := item.toEvent(calID, now)
		if ev.Title == "" {
			continue
		}
		out = append(out, ev)
	}
	return out, nil
}

type rawEvent struct {
	EventID     string `json:"event_id"`
	Summary     string `json:"summary"`
	Description string `json:"description"`
	Status      string `json:"status"`
	Location    struct {
		Name string `json:"name"`
	} `json:"location"`
	StartTime rawTime `json:"start_time"`
	EndTime   rawTime `json:"end_time"`
	Vchat     struct {
		MeetingURL string `json:"meeting_url"`
		VCType     string `json:"vc_type"`
	} `json:"vchat"`
	AppLink string `json:"app_link"`
}

type rawTime struct {
	Timestamp string `json:"timestamp"`
	Date      string `json:"date"`
	Timezone  string `json:"timezone"`
}

func (t rawTime) Time() time.Time {
	if t.Timestamp != "" {
		sec, err := strconv.ParseInt(t.Timestamp, 10, 64)
		if err == nil {
			return time.Unix(sec, 0)
		}
	}
	if t.Date != "" {
		loc := time.Local
		if t.Timezone != "" {
			if l, err := time.LoadLocation(t.Timezone); err == nil {
				loc = l
			}
		}
		if d, err := time.ParseInLocation("2006-01-02", t.Date, loc); err == nil {
			return d
		}
	}
	return time.Time{}
}

func (r rawEvent) toEvent(calID string, now time.Time) Event {
	start := r.StartTime.Time()
	end := r.EndTime.Time()
	hasMeeting := r.Vchat.MeetingURL != "" || r.Vchat.VCType != ""
	return Event{
		ID:          r.EventID,
		CalendarID:  calID,
		Title:       r.Summary,
		Start:       start,
		End:         end,
		Location:    r.Location.Name,
		Description: strings.TrimSpace(stripHTML(r.Description)),
		HasMeeting:  hasMeeting,
		Status:      ClassifyStatus(start, end, now),
		MeetingURL:  r.Vchat.MeetingURL,
		SourceURL:   r.AppLink,
	}
}

func (c *Client) resolveCalendarID(calendarID string) (string, error) {
	if calendarID != "" && calendarID != "primary" {
		return calendarID, nil
	}
	var primary struct {
		Calendars []struct {
			Calendar struct {
				CalendarID string `json:"calendar_id"`
			} `json:"calendar"`
		} `json:"calendars"`
	}
	if err := c.post("/calendar/v4/calendars/primary", map[string]any{}, &primary); err != nil {
		return "", err
	}
	if len(primary.Calendars) == 0 {
		return "", fmt.Errorf("找不到主日历")
	}
	return primary.Calendars[0].Calendar.CalendarID, nil
}

func (c *Client) GetEvent(calendarID, eventID string, now time.Time) (*Event, error) {
	calID, err := c.resolveCalendarID(calendarID)
	if err != nil {
		return nil, err
	}
	var data struct {
		Event rawEvent `json:"event"`
	}
	path := "/calendar/v4/calendars/" + url.PathEscape(calID) + "/events/" + url.PathEscape(eventID)
	if err := c.get(path, nil, &data); err != nil {
		return nil, err
	}
	ev := data.Event.toEvent(calID, now)
	if atts, err := c.listAttendees(calID, eventID); err == nil {
		ev.Attendees = atts
	}
	return &ev, nil
}

func (c *Client) listAttendees(calendarID, eventID string) ([]string, error) {
	q := url.Values{}
	q.Set("page_size", "50")
	var data struct {
		Items []struct {
			DisplayName string `json:"display_name"`
		} `json:"items"`
	}
	path := "/calendar/v4/calendars/" + url.PathEscape(calendarID) + "/events/" + url.PathEscape(eventID) + "/attendees"
	if err := c.get(path, q, &data); err != nil {
		return nil, err
	}
	names := make([]string, 0, len(data.Items))
	for _, it := range data.Items {
		if it.DisplayName != "" {
			names = append(names, it.DisplayName)
		}
	}
	return names, nil
}

func (c *Client) ListOpenTasks(now time.Time) ([]Task, error) {
	q := url.Values{}
	q.Set("page_size", "50")
	q.Set("completed", "false")
	var data struct {
		Items []struct {
			GUID        string `json:"guid"`
			Summary     string `json:"summary"`
			URL         string `json:"url"`
			CompletedAt string `json:"completed_at"`
			Due         struct {
				Timestamp string `json:"timestamp"`
			} `json:"due"`
		} `json:"items"`
	}
	if err := c.get("/task/v2/tasks", q, &data); err != nil {
		return nil, err
	}
	out := make([]Task, 0, len(data.Items))
	for _, it := range data.Items {
		t := Task{ID: it.GUID, Title: it.Summary, URL: it.URL, Done: it.CompletedAt != ""}
		if it.Due.Timestamp != "" {
			if sec, err := strconv.ParseInt(it.Due.Timestamp, 10, 64); err == nil {
				due := time.Unix(sec, 0)
				t.Due = &due
				t.Overdue = due.Before(now) && !t.Done
			}
		}
		out = append(out, t)
	}
	return out, nil
}

func (c *Client) ListRecentDocs(now time.Time) ([]Doc, error) {
	q := url.Values{}
	q.Set("page_size", "30")
	q.Set("order_by", "EditedTime")
	q.Set("direction", "DESC")
	var data struct {
		Files []struct {
			Token        string `json:"token"`
			Name         string `json:"name"`
			Type         string `json:"type"`
			URL          string `json:"url"`
			ModifiedTime any    `json:"modified_time"`
		} `json:"files"`
	}
	if err := c.get("/drive/v1/files", q, &data); err != nil {
		return nil, err
	}
	out := make([]Doc, 0, len(data.Files))
	for _, f := range data.Files {
		if f.Type == "folder" || f.Name == "" {
			continue
		}
		mod := parseFlexTime(f.ModifiedTime)
		days := int(now.Sub(mod).Hours() / 24)
		if days < 0 {
			days = 0
		}
		out = append(out, Doc{
			Token:      f.Token,
			Name:       f.Name,
			Type:       f.Type,
			URL:        f.URL,
			ModifiedAt: mod,
			Stale:      days >= 30,
			DaysIdle:   days,
		})
	}
	return out, nil
}

func (c *Client) ListRecentChats() ([]Chat, error) {
	q := url.Values{}
	q.Set("page_size", "20")
	q.Set("sort_type", "ByActiveTimeDesc")
	var data struct {
		Items []struct {
			ChatID      string `json:"chat_id"`
			Name        string `json:"name"`
			Avatar      string `json:"avatar"`
			External    bool   `json:"external"`
			UserCount any `json:"user_count"`
		} `json:"items"`
	}
	if err := c.get("/im/v1/chats", q, &data); err != nil {
		return nil, err
	}
	out := make([]Chat, 0, len(data.Items))
	for _, it := range data.Items {
		n := anyToInt(it.UserCount)
		out = append(out, Chat{
			ID:          it.ChatID,
			Name:        it.Name,
			Avatar:      it.Avatar,
			External:    it.External,
			MemberCount: n,
			UpdatedAt:   time.Now(),
		})
	}
	return out, nil
}

func (c *Client) CreateDocument(title string, markdown string) (string, error) {
	var created struct {
		Document struct {
			DocumentID string `json:"document_id"`
			URL        string `json:"url"`
		} `json:"document"`
	}
	if err := c.post("/docx/v1/documents", map[string]any{"title": title}, &created); err != nil {
		return "", err
	}
	docID := created.Document.DocumentID
	url := created.Document.URL
	children := markdownToBlocks(markdown)
	if len(children) > 0 {
		_ = c.post("/docx/v1/documents/"+docID+"/blocks/"+docID+"/children", map[string]any{
			"children": children,
		}, nil)
	}
	if url == "" {
		url = "https://feishu.cn/docx/" + docID
	}
	return url, nil
}

func (c *Client) CreateTask(summary string) error {
	return c.post("/task/v2/tasks", map[string]any{"summary": summary}, nil)
}

func (c *Client) CreateOfficialMinute(calendarID, eventID string) (string, error) {
	var data struct {
		DocURL string `json:"doc_url"`
	}
	path := "/calendar/v4/calendars/" + url.PathEscape(calendarID) + "/events/" + url.PathEscape(eventID) + "/meeting_minute"
	if err := c.post(path, map[string]any{}, &data); err != nil {
		return "", err
	}
	return data.DocURL, nil
}

func markdownToBlocks(md string) []map[string]any {
	var blocks []map[string]any
	for _, line := range strings.Split(md, "\n") {
		line = strings.TrimRight(line, "\r")
		if strings.TrimSpace(line) == "" {
			continue
		}
		switch {
		case strings.HasPrefix(line, "# "):
			blocks = append(blocks, headingBlock(3, strings.TrimPrefix(line, "# ")))
		case strings.HasPrefix(line, "## "):
			blocks = append(blocks, headingBlock(4, strings.TrimPrefix(line, "## ")))
		case strings.HasPrefix(line, "### "):
			blocks = append(blocks, headingBlock(5, strings.TrimPrefix(line, "### ")))
		case strings.HasPrefix(line, "- "):
			blocks = append(blocks, listBlock(12, strings.TrimPrefix(line, "- ")))
		default:
			blocks = append(blocks, textBlock(line))
		}
	}
	return blocks
}

func headingBlock(typ int, text string) map[string]any {
	key := "heading1"
	switch typ {
	case 4:
		key = "heading2"
	case 5:
		key = "heading3"
	}
	return map[string]any{
		"block_type": typ,
		key: map[string]any{
			"elements": []map[string]any{{"text_run": map[string]any{"content": text}}},
		},
	}
}

func textBlock(text string) map[string]any {
	return map[string]any{
		"block_type": 2,
		"text": map[string]any{
			"elements": []map[string]any{{"text_run": map[string]any{"content": text}}},
		},
	}
}

func listBlock(typ int, text string) map[string]any {
	key := "bullet"
	if typ == 13 {
		key = "ordered"
	}
	return map[string]any{
		"block_type": typ,
		key: map[string]any{
			"elements": []map[string]any{{"text_run": map[string]any{"content": text}}},
		},
	}
}

func stripHTML(s string) string {
	var b strings.Builder
	in := false
	for _, r := range s {
		switch r {
		case '<':
			in = true
		case '>':
			in = false
		default:
			if !in {
				b.WriteRune(r)
			}
		}
	}
	return b.String()
}

func parseFlexTime(v any) time.Time {
	switch t := v.(type) {
	case float64:
		n := int64(t)
		if n > 1e12 {
			return time.UnixMilli(n)
		}
		if n > 0 {
			return time.Unix(n, 0)
		}
	case string:
		if t == "" {
			return time.Time{}
		}
		if sec, err := strconv.ParseInt(t, 10, 64); err == nil {
			if sec > 1e12 {
				return time.UnixMilli(sec)
			}
			return time.Unix(sec, 0)
		}
	}
	return time.Time{}
}

func anyToInt(v any) int {
	switch t := v.(type) {
	case float64:
		return int(t)
	case string:
		n, _ := strconv.Atoi(t)
		return n
	}
	return 0
}

func friendlyErr(err error) string {
	if IsPermission(err) {
		return "当前账号未授权此项权限，已跳过"
	}
	return err.Error()
}
