package feishu

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

const (
	AuthorizeURL = "https://accounts.feishu.cn/open-apis/authen/v1/authorize"
	TokenURL     = "https://open.feishu.cn/open-apis/authen/v2/oauth/token"
	UserInfoURL  = "https://open.feishu.cn/open-apis/authen/v1/user_info"
)

// 只申请雷达和纪要真正用到的权限，写回文档/任务需要 write。
const DefaultScopes = "offline_access auth:user.id:read calendar:calendar:readonly calendar:calendar drive:drive:readonly drive:drive docx:document:readonly docx:document task:task:read task:task:write im:chat:readonly"

type TokenSet struct {
	AccessToken           string
	RefreshToken          string
	ExpiresIn             int
	RefreshExpiresIn      int
	Scope                 string
}

type UserInfo struct {
	Name      string `json:"name"`
	OpenID    string `json:"open_id"`
	UnionID   string `json:"union_id"`
	Avatar    string `json:"avatar_url"`
	Email     string `json:"email"`
	TenantKey string `json:"tenant_key"`
}

type stateEntry struct {
	at time.Time
}

var (
	stateMu sync.Mutex
	states  = map[string]stateEntry{}
)

func NewState() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	s := hex.EncodeToString(b)
	stateMu.Lock()
	states[s] = stateEntry{at: time.Now()}
	stateMu.Unlock()
	return s
}

func ConsumeState(s string) bool {
	if s == "" {
		return false
	}
	stateMu.Lock()
	defer stateMu.Unlock()
	ent, ok := states[s]
	delete(states, s)
	if !ok {
		return false
	}
	return time.Since(ent.at) < 10*time.Minute
}

func AuthorizeLink(appID, redirectURI, state string) string {
	q := url.Values{}
	q.Set("client_id", appID)
	q.Set("response_type", "code")
	q.Set("redirect_uri", redirectURI)
	q.Set("scope", DefaultScopes)
	q.Set("state", state)
	q.Set("prompt", "consent")
	return AuthorizeURL + "?" + q.Encode()
}

func ExchangeCode(appID, appSecret, code, redirectURI string) (*TokenSet, error) {
	return postToken(map[string]string{
		"grant_type":    "authorization_code",
		"client_id":     appID,
		"client_secret": appSecret,
		"code":          code,
		"redirect_uri":  redirectURI,
	})
}

func Refresh(appID, appSecret, refreshToken string) (*TokenSet, error) {
	return postToken(map[string]string{
		"grant_type":    "refresh_token",
		"client_id":     appID,
		"client_secret": appSecret,
		"refresh_token": refreshToken,
	})
}

func postToken(body map[string]string) (*TokenSet, error) {
	raw, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, TokenURL, bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var out struct {
		Code                  int    `json:"code"`
		Error                 string `json:"error"`
		ErrorDescription      string `json:"error_description"`
		AccessToken           string `json:"access_token"`
		RefreshToken          string `json:"refresh_token"`
		ExpiresIn             int    `json:"expires_in"`
		RefreshTokenExpiresIn int    `json:"refresh_token_expires_in"`
		Scope                 string `json:"scope"`
		Msg                   string `json:"msg"`
	}
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, fmt.Errorf("兑换 token 失败: %s", string(data))
	}
	if out.Code != 0 || out.AccessToken == "" {
		msg := out.Msg
		if msg == "" {
			msg = strings.TrimSpace(out.Error + " " + out.ErrorDescription)
		}
		if msg == "" {
			msg = string(data)
		}
		return nil, fmt.Errorf("兑换 token 失败: %s", msg)
	}
	return &TokenSet{
		AccessToken:      out.AccessToken,
		RefreshToken:     out.RefreshToken,
		ExpiresIn:        out.ExpiresIn,
		RefreshExpiresIn: out.RefreshTokenExpiresIn,
		Scope:            out.Scope,
	}, nil
}

func FetchUserInfo(accessToken string) (*UserInfo, error) {
	req, err := http.NewRequest(http.MethodGet, UserInfoURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var out struct {
		Code int      `json:"code"`
		Msg  string   `json:"msg"`
		Data UserInfo `json:"data"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, err
	}
	if out.Code != 0 {
		return nil, fmt.Errorf("获取用户信息失败: %s", out.Msg)
	}
	if out.Data.Avatar == "" {
		var alt struct {
			Data struct {
				AvatarURL string `json:"avatar_url"`
				Avatar    struct {
					Avatar72 string `json:"avatar_72"`
				} `json:"avatar"`
			} `json:"data"`
		}
		_ = json.Unmarshal(raw, &alt)
		if alt.Data.AvatarURL != "" {
			out.Data.Avatar = alt.Data.AvatarURL
		} else {
			out.Data.Avatar = alt.Data.Avatar.Avatar72
		}
	}
	return &out.Data, nil
}
