package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"

	"multi-model-rag-platform/internal/config"
)

type DemoService struct {
	config config.Config
}

func NewDemoService(cfg config.Config) DemoService {
	return DemoService{config: cfg}
}

func (s DemoService) UnlockEnabled() bool {
	return s.config.DemoRealModePassword != ""
}

func (s DemoService) IsUnlocked(r *http.Request) bool {
	if !s.UnlockEnabled() {
		return false
	}
	cookie, err := r.Cookie(s.config.DemoUnlockCookieName)
	if err != nil || cookie.Value == "" {
		return false
	}
	return hmac.Equal([]byte(cookie.Value), []byte(s.expectedCookieValue()))
}

func (s DemoService) ValidatePassword(password string) bool {
	if !s.UnlockEnabled() {
		return false
	}
	return hmac.Equal([]byte(password), []byte(s.config.DemoRealModePassword))
}

func (s DemoService) SetUnlockCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     s.config.DemoUnlockCookieName,
		Value:    s.expectedCookieValue(),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
		MaxAge:   60 * 60 * 8,
		Path:     "/",
	})
}

func (s DemoService) ClearUnlockCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     s.config.DemoUnlockCookieName,
		Value:    "",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
		MaxAge:   -1,
		Path:     "/",
	})
}

func (s DemoService) expectedCookieValue() string {
	secret := s.config.DemoUnlockCookieSecret
	if secret == "" {
		secret = s.config.DemoRealModePassword
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte("demo-real-mode-unlocked"))
	return "v1:" + hex.EncodeToString(mac.Sum(nil))
}
