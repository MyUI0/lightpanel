package handlers

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"lightpanel/config"
	"lightpanel/models"
)

var sessionDir string
var accountMu sync.Mutex

func init() {
	sessionDir = config.ConfigDir + "/sessions"
	maxLoginAttempts = 1000
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			loginAttemptsMu.Lock()
			now := time.Now()
			for ip, a := range loginAttempts {
				if !a.Locked && now.Sub(a.LastFail) > 15*time.Minute {
					delete(loginAttempts, ip)
				} else if a.Locked && now.After(a.LockUntil) {
					delete(loginAttempts, ip)
				}
			}
			loginAttemptsMu.Unlock()
		}
	}()
}

func hash(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}

// 登录限流
var (
	loginAttempts    = make(map[string]*loginAttempt)
	loginAttemptsMu  sync.Mutex
	maxLoginAttempts int
)

type loginAttempt struct {
	Count     int
	LastFail  time.Time
	Locked    bool
	LockUntil time.Time
}

func checkLoginLimit(ip string) bool {
	loginAttemptsMu.Lock()
	defer loginAttemptsMu.Unlock()

	a, ok := loginAttempts[ip]
	if !ok {
		if len(loginAttempts) >= maxLoginAttempts {
			return false
		}
		loginAttempts[ip] = &loginAttempt{}
		return true
	}

	if a.Locked && time.Now().Before(a.LockUntil) {
		return false
	}

	if a.Locked && time.Now().After(a.LockUntil) {
		a.Count = 0
		a.Locked = false
	}

	if time.Since(a.LastFail) > 15*time.Minute {
		a.Count = 0
		a.Locked = false
	}

	return true
}

func recordLoginFail(ip string) {
	loginAttemptsMu.Lock()
	defer loginAttemptsMu.Unlock()

	a, ok := loginAttempts[ip]
	if !ok {
		a = &loginAttempt{}
		loginAttempts[ip] = a
	}

	a.Count++
	a.LastFail = time.Now()

	if a.Count >= 5 {
		a.Locked = true
		a.LockUntil = time.Now().Add(15 * time.Minute)
	}
}

func recordLoginSuccess(ip string) {
	loginAttemptsMu.Lock()
	defer loginAttemptsMu.Unlock()
	delete(loginAttempts, ip)
}

func getLoginRemainTime(ip string) int {
	loginAttemptsMu.Lock()
	defer loginAttemptsMu.Unlock()

	a, ok := loginAttempts[ip]
	if !ok || !a.Locked {
		return 0
	}
	remain := int(time.Until(a.LockUntil).Seconds())
	if remain < 0 {
		return 0
	}
	return remain
}

// SSRF 防护：检查是否为内网地址（定义见 security.go）

func generateSessionToken() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return hex.EncodeToString(b)
}

func isValidSession(token string) bool { return getSessionData(token) != nil }

type sessionData struct {
	Credential string `json:"credential"`
	Expires    int64  `json:"expires"`
	CSRFToken  string `json:"csrf"`
	FirstLogin bool   `json:"first_login"`
}

func getSessionData(token string) *sessionData {
	if token == "" || len(token) != 64 {
		return nil
	}
	if _, err := hex.DecodeString(token); err != nil {
		return nil
	}
	sessionFile := filepath.Join(sessionDir, filepath.Base(token)+".json")
	b, err := os.ReadFile(sessionFile)
	if err != nil {
		return nil
	}
	var sess sessionData
	if err := json.Unmarshal(b, &sess); err != nil {
		return nil
	}
	if time.Now().Unix() > sess.Expires {
		_ = os.Remove(sessionFile)
		return nil
	}
	var user models.UserConfig
	if LoadJSON(config.ConfigUsr, &user) != nil || sess.Credential != credentialDigest(user) {
		return nil
	}
	return &sess
}

func isFirstLogin(r *http.Request) bool {
	cookie, _ := r.Cookie("lp_session")
	if cookie == nil {
		return false
	}
	sessData := getSessionData(cookie.Value)
	return sessData != nil && sessData.FirstLogin
}

func markPasswordChanged(w http.ResponseWriter, r *http.Request) {
	cookie, _ := r.Cookie("lp_session")
	if cookie == nil {
		return
	}
	sessData := getSessionData(cookie.Value)
	if sessData == nil {
		return
	}
	sessionFile := filepath.Join(sessionDir, filepath.Base(cookie.Value)+".json")
	sessData.FirstLogin = false
	data, _ := json.Marshal(sessData)
	_ = os.WriteFile(sessionFile, data, 0600)
}

func createSession(w http.ResponseWriter, firstLogin bool, secure ...bool) string {
	var user models.UserConfig
	if LoadJSON(config.ConfigUsr, &user) != nil {
		return ""
	}
	secureCookie := os.Getenv("LIGHTPANEL_SECURE_COOKIES") == "1" || (len(secure) > 0 && secure[0])
	token := generateSessionToken()
	if token == "" {
		return ""
	}
	csrfToken := generateSessionToken()
	if csrfToken == "" {
		return ""
	}
	if err := os.MkdirAll(sessionDir, 0700); err != nil {
		return ""
	}
	sessData, _ := json.Marshal(sessionData{
		Credential: credentialDigest(user),
		Expires:    time.Now().Add(24 * time.Hour).Unix(),
		CSRFToken:  csrfToken,
		FirstLogin: firstLogin,
	})
	if err := os.WriteFile(filepath.Join(sessionDir, token+".json"), sessData, 0600); err != nil {
		return ""
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "lp_session",
		Value:    token,
		Path:     "/",
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Secure:   secureCookie,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "lp_csrf",
		Value:    csrfToken,
		Path:     "/",
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: false,
		SameSite: http.SameSiteStrictMode,
		Secure:   secureCookie,
	})
	return token
}

func destroySession(token string) {
	if token != "" {
		_ = os.Remove(filepath.Join(sessionDir, filepath.Base(token)+".json"))
	}
}

func CleanupSessions() {
	files, _ := os.ReadDir(sessionDir)
	for _, f := range files {
		b, err := os.ReadFile(filepath.Join(sessionDir, f.Name()))
		if err != nil {
			continue
		}
		var sess sessionData
		if json.Unmarshal(b, &sess) == nil && time.Now().Unix() > sess.Expires {
			_ = os.Remove(filepath.Join(sessionDir, f.Name()))
		}
	}
}

func auth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, _ := r.Cookie("lp_session")
		if cookie != nil && isValidSession(cookie.Value) {
			if isFirstLogin(r) && r.URL.Path != "/setting" {
				http.Redirect(w, r, "/setting?first=1", 303)
				return
			}
			next(w, r)
			return
		}
		http.Redirect(w, r, "/login", 302)
	}
}

func authWithCSRF(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, _ := r.Cookie("lp_session")
		if cookie == nil || !isValidSession(cookie.Value) {
			http.Redirect(w, r, "/login", 302)
			return
		}

		if isFirstLogin(r) && r.URL.Path != "/setting/save-account" {
			http.Error(w, "请先修改初始密码", 403)
			return
		}
		if r.Method != http.MethodPost {
			w.Header().Set("Allow", "POST")
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}
		r.Body = http.MaxBytesReader(w, r.Body, int64(config.MaxDownBytes)+1024*1024)
		defer func() {
			if r.MultipartForm != nil {
				_ = r.MultipartForm.RemoveAll()
			}
		}()
		csrfToken := r.Header.Get("X-CSRF-Token")
		if csrfToken == "" {
			csrfToken = r.FormValue("csrf_token")
		}

		sessData := getSessionData(cookie.Value)
		if sessData == nil || csrfToken == "" || csrfToken != sessData.CSRFToken {
			http.Error(w, "CSRF validation failed", http.StatusForbidden)
			return
		}
		if strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
			if err := r.ParseMultipartForm(16 << 20); err != nil {
				http.Error(w, "表单格式或大小无效", 400)
				return
			}
		} else if err := r.ParseForm(); err != nil {
			http.Error(w, "表单格式无效", 400)
			return
		}
		next(w, r)
	}
}

func loginPage(w http.ResponseWriter, r *http.Request) {
	err := r.URL.Query().Get("err")
	lockTime := getLoginRemainTime(getClientIP(r))
	_ = htmlRender.ExecuteTemplate(w, "login", map[string]any{
		"Err":      err,
		"LockTime": lockTime,
	})
}

func getClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); os.Getenv("LIGHTPANEL_TRUST_PROXY") == "1" && xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}
	if xri := r.Header.Get("X-Real-IP"); os.Getenv("LIGHTPANEL_TRUST_PROXY") == "1" && xri != "" {
		return xri
	}
	host, _, _ := net.SplitHostPort(r.RemoteAddr)
	if host != "" {
		return host
	}
	return r.RemoteAddr
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 8192)
	if err := r.ParseForm(); err != nil {
		http.Error(w, "登录请求过大或无效", 400)
		return
	}
	accountMu.Lock()
	defer accountMu.Unlock()
	user := r.Form.Get("username")
	pass := r.Form.Get("password")

	ip := getClientIP(r)

	if !checkLoginLimit(ip) {
		http.Redirect(w, r, "/login?err=locked", 302)
		return
	}

	var usr models.UserConfig
	if err := LoadJSON(config.ConfigUsr, &usr); err != nil {
		http.Redirect(w, r, "/login?err=1", 302)
		return
	}

	if user == usr.Username && passwordMatches(usr.Password, pass) {
		recordLoginSuccess(ip)
		isDefaultPassword := accountNeedsChange(usr)
		if len(usr.Password) == 64 {
			upgraded, err := passwordHash(pass)
			if err != nil {
				http.Error(w, "凭据升级失败", 500)
				return
			}
			usr.Password = upgraded
			usr.MustChangePassword = isDefaultPassword
			if err = WriteJSON(config.ConfigUsr, usr); err != nil {
				http.Error(w, "凭据保存失败", 500)
				return
			}
		}
		if createSession(w, isDefaultPassword, r.TLS != nil) == "" {
			http.Error(w, "无法创建会话", 500)
			return
		}
		http.Redirect(w, r, "/", 302)
		return
	}

	recordLoginFail(ip)
	remain := getLoginRemainTime(ip)
	if remain > 0 {
		http.Redirect(w, r, "/login?err=locked", 302)
	} else {
		http.Redirect(w, r, "/login?err=1", 302)
	}
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	cookie, _ := r.Cookie("lp_session")
	if cookie != nil {
		destroySession(cookie.Value)
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "lp_session",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
	http.Redirect(w, r, "/login", 302)
}

func requirePOST(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.WriteHeader(405)
			_, _ = w.Write([]byte("Method Not Allowed"))
			return
		}
		next(w, r)
	}
}

func InitConfig() error {
	sessionDir = filepath.Join(config.ConfigDir, "sessions")
	if err := EnsureDir(config.ConfigDir); err != nil {
		return err
	}
	if err := EnsureDir(config.AppsDir); err != nil {
		return err
	}

	if _, e := os.Stat(config.ConfigApps); e != nil {
		if err := WriteJSON(config.ConfigApps, map[string]models.Project{}); err != nil {
			return err
		}
	}
	if _, e := os.Stat(config.ConfigSrc); e != nil {
		if err := WriteJSON(config.ConfigSrc, []models.StoreSource{}); err != nil {
			return err
		}
	}
	if _, e := os.Stat(config.ConfigUsr); e != nil {
		password := os.Getenv("LIGHTPANEL_INITIAL_PASSWORD")
		if password == "" {
			password = generateSessionToken()
			if len(password) < 24 {
				return fmt.Errorf("无法生成初始密码")
			}
			password = password[:24]
		}
		if len(password) < 12 || len(password) > 1024 {
			return fmt.Errorf("初始密码长度必须为12至1024字节")
		}
		encoded, err := passwordHash(password)
		if err != nil {
			return err
		}
		if err = os.WriteFile(filepath.Join(config.ConfigDir, "initial-password.txt"), []byte(password+"\n"), 0600); err != nil {
			return err
		}
		if err = WriteJSON(config.ConfigUsr, models.UserConfig{Username: "admin", Password: encoded, MustChangePassword: true}); err != nil {
			return err
		}
		log.Printf("首次登录账号：admin；初始密码保存在 %s，登录后必须修改", filepath.Join(config.ConfigDir, "initial-password.txt"))
	}
	if _, e := os.Stat(config.ConfigBg); e != nil {
		if err := WriteJSON(config.ConfigBg, models.BgConfig{Type: "gradient", URL: ""}); err != nil {
			return err
		}
	}
	return nil
}
