package skills

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"feishushow/config"
	"feishushow/feishu"
)

func BuildMinutes(cfg *config.Config, ev feishu.Event, now time.Time) feishu.MinutesResult {
	if cfg.LLMConfigured() {
		if res, err := fromLLM(cfg, ev, now); err == nil {
			res.EventID = ev.ID
			res.Title = "会议纪要 · " + ev.Title
			res.Source = "llm"
			return res
		}
	}
	return fromTemplate(ev, now)
}

func fromTemplate(ev feishu.Event, now time.Time) feishu.MinutesResult {
	attendees := ev.Attendees
	if len(attendees) == 0 {
		attendees = []string{"与会同学"}
	}
	owner := attendees[0]
	decisions := []string{
		"本周先交付「今日雷达」一屏，不扩做成通用 Agent。",
		"纪要默认写回飞书文档，待办同步到任务中心。",
	}
	actions := []feishu.Action{
		{Title: "整理本次讨论结论并同步到项目群", Owner: owner, DueHint: "今天"},
		{Title: "补齐未决问题的负责人", Owner: last(attendees), DueHint: "明天"},
	}
	risks := []string{}
	if ev.Description != "" {
		risks = append(risks, "会前材料只给了摘要，细节还需对照原始文档。")
	}
	if ev.Status == "upcoming" {
		risks = append(risks, "会议尚未开始，当前纪要是按日程信息预生成的提纲，会后再改一版。")
	}

	var b strings.Builder
	fmt.Fprintf(&b, "# 会议纪要 · %s\n\n", ev.Title)
	fmt.Fprintf(&b, "- 时间：%s – %s\n", ev.Start.In(now.Location()).Format("01-02 15:04"), ev.End.In(now.Location()).Format("15:04"))
	if ev.Location != "" {
		fmt.Fprintf(&b, "- 地点：%s\n", ev.Location)
	}
	fmt.Fprintf(&b, "- 参与人：%s\n\n", strings.Join(attendees, "、"))
	if ev.Description != "" {
		fmt.Fprintf(&b, "## 会前背景\n\n%s\n\n", ev.Description)
	}
	b.WriteString("## 决议\n\n")
	for _, d := range decisions {
		fmt.Fprintf(&b, "- %s\n", d)
	}
	b.WriteString("\n## 待办\n\n")
	for _, a := range actions {
		fmt.Fprintf(&b, "- [%s] %s（%s）\n", a.Owner, a.Title, a.DueHint)
	}
	if len(risks) > 0 {
		b.WriteString("\n## 风险\n\n")
		for _, r := range risks {
			fmt.Fprintf(&b, "- %s\n", r)
		}
	}
	b.WriteString("\n---\n由 FeishuShow 本地模板生成。配置 DOUBAO_API_KEY 后可改用模型润色。\n")

	return feishu.MinutesResult{
		EventID:   ev.ID,
		Title:     "会议纪要 · " + ev.Title,
		Markdown:  b.String(),
		Decisions: decisions,
		Actions:   actions,
		Risks:     risks,
		Source:    "template",
	}
}

func fromLLM(cfg *config.Config, ev feishu.Event, now time.Time) (feishu.MinutesResult, error) {
	prompt := fmt.Sprintf(`你是飞书会议纪要助手。根据日程信息生成结构化纪要。
只输出 JSON，不要 Markdown 围栏，字段：
{"decisions":["..."],"actions":[{"title":"...","owner":"...","dueHint":"今天|明天|本周"}],"risks":["..."],"summary":"一段话"}

会议：%s
时间：%s - %s
地点：%s
参与人：%s
描述：%s
现在：%s
状态：%s`, ev.Title, ev.Start.Format(time.RFC3339), ev.End.Format(time.RFC3339), ev.Location,
		strings.Join(ev.Attendees, "、"), ev.Description, now.Format(time.RFC3339), ev.Status)

	body := map[string]any{
		"model": cfg.DoubaoModel,
		"messages": []map[string]string{
			{"role": "system", "content": "你把会议收成可执行的纪要，决议短、待办带负责人。"},
			{"role": "user", "content": prompt},
		},
		"temperature": 0.3,
	}
	raw, _ := json.Marshal(body)
	req, err := http.NewRequest(http.MethodPost, cfg.DoubaoBaseURL+"/chat/completions", bytes.NewReader(raw))
	if err != nil {
		return feishu.MinutesResult{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+cfg.DoubaoAPIKey)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return feishu.MinutesResult{}, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return feishu.MinutesResult{}, err
	}
	var out struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(data, &out); err != nil {
		return feishu.MinutesResult{}, err
	}
	if len(out.Choices) == 0 {
		if out.Error.Message != "" {
			return feishu.MinutesResult{}, fmt.Errorf(out.Error.Message)
		}
		return feishu.MinutesResult{}, fmt.Errorf("模型无返回")
	}
	content := strings.TrimSpace(out.Choices[0].Message.Content)
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	var parsed struct {
		Decisions []string        `json:"decisions"`
		Actions   []feishu.Action `json:"actions"`
		Risks     []string        `json:"risks"`
		Summary   string          `json:"summary"`
	}
	if err := json.Unmarshal([]byte(content), &parsed); err != nil {
		return feishu.MinutesResult{}, err
	}
	if len(parsed.Decisions) == 0 {
		return feishu.MinutesResult{}, fmt.Errorf("模型未给出决议")
	}

	var b strings.Builder
	fmt.Fprintf(&b, "# 会议纪要 · %s\n\n", ev.Title)
	if parsed.Summary != "" {
		fmt.Fprintf(&b, "%s\n\n", parsed.Summary)
	}
	fmt.Fprintf(&b, "- 时间：%s – %s\n", ev.Start.In(now.Location()).Format("01-02 15:04"), ev.End.In(now.Location()).Format("15:04"))
	if len(ev.Attendees) > 0 {
		fmt.Fprintf(&b, "- 参与人：%s\n", strings.Join(ev.Attendees, "、"))
	}
	b.WriteString("\n## 决议\n\n")
	for _, d := range parsed.Decisions {
		fmt.Fprintf(&b, "- %s\n", d)
	}
	b.WriteString("\n## 待办\n\n")
	for _, a := range parsed.Actions {
		fmt.Fprintf(&b, "- [%s] %s（%s）\n", a.Owner, a.Title, a.DueHint)
	}
	if len(parsed.Risks) > 0 {
		b.WriteString("\n## 风险\n\n")
		for _, r := range parsed.Risks {
			fmt.Fprintf(&b, "- %s\n", r)
		}
	}

	return feishu.MinutesResult{
		Markdown:  b.String(),
		Decisions: parsed.Decisions,
		Actions:   parsed.Actions,
		Risks:     parsed.Risks,
	}, nil
}

func last(ss []string) string {
	if len(ss) == 0 {
		return "待指定"
	}
	return ss[len(ss)-1]
}
