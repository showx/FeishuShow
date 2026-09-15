package config

import (
	"bufio"
	"os"
	"strings"
	"time"
)

type Config struct {
	Port             string
	JWTSecret        []byte
	JWTExpire        time.Duration
	DBPath           string
	FrontendURL      string
	FeishuAppID      string
	FeishuAppSecret  string
	FeishuRedirect   string
	DoubaoAPIKey     string
	DoubaoBaseURL    string
	DoubaoModel      string
}

func Load() *Config {
	loadDotEnv(".env")
	loadDotEnv("../.env")

	return &Config{
		Port:            getenv("PORT", "8080"),
		JWTSecret:       []byte(getenv("JWT_SECRET", "feishushow-dev-secret-change-in-prod")),
		JWTExpire:       7 * 24 * time.Hour,
		DBPath:          getenv("DB_PATH", "data/feishushow.db"),
		FrontendURL:     getenv("FRONTEND_URL", "http://localhost:5173"),
		FeishuAppID:     getenv("FEISHU_APP_ID", ""),
		FeishuAppSecret: getenv("FEISHU_APP_SECRET", ""),
		FeishuRedirect:  getenv("FEISHU_REDIRECT_URI", "http://localhost:5173/callback"),
		DoubaoAPIKey:    getenv("DOUBAO_API_KEY", ""),
		DoubaoBaseURL:   strings.TrimRight(getenv("DOUBAO_BASE_URL", "https://ark.cn-beijing.volces.com/api/v3"), "/"),
		DoubaoModel:     getenv("DOUBAO_MODEL", "doubao-seed-1-6-250615"),
	}
}

func (c *Config) FeishuConfigured() bool {
	return c.FeishuAppID != "" && c.FeishuAppSecret != ""
}

func (c *Config) LLMConfigured() bool {
	return c.DoubaoAPIKey != ""
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func loadDotEnv(path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, val, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		val = strings.TrimSpace(val)
		if strings.HasPrefix(val, `"`) && strings.HasSuffix(val, `"`) && len(val) >= 2 {
			val = val[1 : len(val)-1]
		}
		if os.Getenv(key) == "" {
			_ = os.Setenv(key, val)
		}
	}
}
