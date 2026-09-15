package feishu

import "time"

type Event struct {
	ID           string    `json:"id"`
	CalendarID   string    `json:"calendarId"`
	Title        string    `json:"title"`
	Start        time.Time `json:"start"`
	End          time.Time `json:"end"`
	Location     string    `json:"location"`
	Description  string    `json:"description"`
	Attendees    []string  `json:"attendees"`
	HasMeeting   bool      `json:"hasMeeting"`
	Status       string    `json:"status"` // upcoming | live | ended
	MeetingURL   string    `json:"meetingUrl"`
	SourceURL    string    `json:"sourceUrl"`
}

type Task struct {
	ID      string     `json:"id"`
	Title   string     `json:"title"`
	Due     *time.Time `json:"due"`
	Done    bool       `json:"done"`
	URL     string     `json:"url"`
	Overdue bool       `json:"overdue"`
}

type Doc struct {
	Token      string    `json:"token"`
	Name       string    `json:"name"`
	Type       string    `json:"type"`
	URL        string    `json:"url"`
	ModifiedAt time.Time `json:"modifiedAt"`
	Stale      bool      `json:"stale"`
	DaysIdle   int       `json:"daysIdle"`
}

type Chat struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	Avatar     string    `json:"avatar"`
	UpdatedAt  time.Time `json:"updatedAt"`
	External   bool      `json:"external"`
	MemberCount int      `json:"memberCount"`
}

type TodayBoard struct {
	Date     string   `json:"date"`
	Stats    Stats    `json:"stats"`
	Events   []Event  `json:"events"`
	Tasks    []Task   `json:"tasks"`
	Docs     []Doc    `json:"docs"`
	Chats    []Chat   `json:"chats"`
	Warnings []string `json:"warnings"`
}

type Stats struct {
	Meetings  int `json:"meetings"`
	Live      int `json:"live"`
	Ended     int `json:"ended"`
	Tasks     int `json:"tasks"`
	Overdue   int `json:"overdue"`
	StaleDocs int `json:"staleDocs"`
	Chats     int `json:"chats"`
}

type MinutesResult struct {
	EventID   string   `json:"eventId"`
	Title     string   `json:"title"`
	Markdown  string   `json:"markdown"`
	Decisions []string `json:"decisions"`
	Actions   []Action `json:"actions"`
	Risks     []string `json:"risks"`
	DocURL    string   `json:"docUrl"`
	Source    string   `json:"source"` // template | llm | feishu
}

type Action struct {
	Title    string `json:"title"`
	Owner    string `json:"owner"`
	DueHint  string `json:"dueHint"`
}

func ClassifyStatus(start, end, now time.Time) string {
	if now.After(end) {
		return "ended"
	}
	if !now.Before(start) && !now.After(end) {
		return "live"
	}
	return "upcoming"
}
