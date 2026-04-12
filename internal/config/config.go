package config

import (
	"os"
	"strings"
)

type Config struct {
	Port                   string
	LogLevel               string
	EnableTracing          bool
	CORSAllowOrigins       []string
	DemoRealModePassword   string
	DemoUnlockCookieSecret string
	DemoUnlockCookieName   string
}

func Load() Config {
	return Config{
		Port:                   envOrDefault("PORT", "8080"),
		LogLevel:               envOrDefault("LOG_LEVEL", "info"),
		EnableTracing:          strings.EqualFold(envOrDefault("ENABLE_TRACING", "true"), "true"),
		CORSAllowOrigins:       parseCSVEnv("CORS_ALLOW_ORIGINS", []string{"http://localhost:5173", "http://127.0.0.1:5173"}),
		DemoRealModePassword:   envOrDefault("DEMO_REAL_MODE_PASSWORD", ""),
		DemoUnlockCookieSecret: envOrDefault("DEMO_UNLOCK_COOKIE_SECRET", ""),
		DemoUnlockCookieName:   envOrDefault("DEMO_UNLOCK_COOKIE_NAME", "mmrag_demo_unlock"),
	}
}

func envOrDefault(name string, fallback string) string {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return fallback
	}
	return value
}

func parseCSVEnv(name string, fallback []string) []string {
	raw := strings.TrimSpace(os.Getenv(name))
	if raw == "" {
		return append([]string(nil), fallback...)
	}

	items := strings.Split(raw, ",")
	values := make([]string, 0, len(items))
	for _, item := range items {
		trimmed := strings.TrimSpace(item)
		if trimmed != "" {
			values = append(values, trimmed)
		}
	}
	if len(values) == 0 {
		return append([]string(nil), fallback...)
	}
	return values
}
