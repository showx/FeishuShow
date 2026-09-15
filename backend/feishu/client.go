package feishu

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const APIBase = "https://open.feishu.cn/open-apis"

type Client struct {
	http *http.Client
	token string
}

func NewClient(token string) *Client {
	return &Client{
		http:  &http.Client{Timeout: 20 * time.Second},
		token: token,
	}
}

func (c *Client) get(path string, query url.Values, dest any) error {
	u := APIBase + path
	if len(query) > 0 {
		u += "?" + query.Encode()
	}
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return err
	}
	return c.do(req, dest)
}

func (c *Client) post(path string, body any, dest any) error {
	var rdr io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return err
		}
		rdr = bytes.NewReader(b)
	}
	req, err := http.NewRequest(http.MethodPost, APIBase+path, rdr)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	return c.do(req, dest)
}

func (c *Client) do(req *http.Request, dest any) error {
	req.Header.Set("Authorization", "Bearer "+c.token)
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	var envelope struct {
		Code int             `json:"code"`
		Msg  string          `json:"msg"`
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(raw, &envelope); err != nil {
		return fmt.Errorf("飞书响应无法解析: %s", truncate(string(raw), 200))
	}
	if envelope.Code != 0 {
		return fmt.Errorf("飞书 API %d: %s", envelope.Code, envelope.Msg)
	}
	if dest != nil && len(envelope.Data) > 0 {
		return json.Unmarshal(envelope.Data, dest)
	}
	return nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

func IsPermission(err error) bool {
	if err == nil {
		return false
	}
	s := err.Error()
	return strings.Contains(s, "99991679") ||
		strings.Contains(s, "99991672") ||
		strings.Contains(s, "permission") ||
		strings.Contains(s, "Unauthorized")
}
