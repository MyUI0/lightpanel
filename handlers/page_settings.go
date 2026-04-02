package handlers

import (
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"lightpanel/config"
	"lightpanel/models"
)

func settingPage(w http.ResponseWriter, r *http.Request) {
	var usr models.UserConfig
	_ = LoadJSON(config.ConfigUsr, &usr)

	var bgConfig models.BgConfig
	_ = LoadJSON(config.ConfigBg, &bgConfig)

	msg := r.URL.Query().Get("ok")
	errMsg := r.URL.Query().Get("err")

	_ = htmlRender.ExecuteTemplate(w, "setting", map[string]any{
		"User":   usr,
		"BgUrl":  bgConfig.URL,
		"BgType": bgConfig.Type,
		"Msg":    msg,
		"Err":    errMsg,
	})
}

func saveSetting(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()

	oldPwd := strings.TrimSpace(r.Form.Get("old"))
	newPwd := strings.TrimSpace(r.Form.Get("new"))
	bgUrl := strings.TrimSpace(r.Form.Get("bg_url"))
	bgType := strings.TrimSpace(r.Form.Get("bg_type"))

	if oldPwd != "" || newPwd != "" {
		var usr models.UserConfig
		_ = LoadJSON(config.ConfigUsr, &usr)
		if oldPwd != usr.Password {
			http.Redirect(w, r, "/setting?err=password", 302)
			return
		}
		if newPwd != "" {
			usr.Password = newPwd
			_ = WriteJSON(config.ConfigUsr, usr)
		}
	}

	if bgType != "" {
		bg := models.BgConfig{Type: bgType, URL: bgUrl}
		_ = WriteJSON(config.ConfigBg, bg)
	}

	http.Redirect(w, r, "/setting?ok=1", 302)
}

func logPage(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/log/")
	var apps map[string]models.Project
	_ = LoadJSON(config.ConfigApps, &apps)
	app, ok := apps[name]
	if !ok {
		http.Redirect(w, r, "/", 302)
		return
	}
	logFile := filepath.Join(app.Path, "run.log")
	content := ""
	lineCount := 0
	if b, e := os.ReadFile(logFile); e == nil {
		if len(b) > config.MaxLogLen {
			content = string(b[len(b)-config.MaxLogLen:])
		} else {
			content = string(b)
		}
		lineCount = strings.Count(content, "\n")
	}
	_ = htmlRender.ExecuteTemplate(w, "log", map[string]any{"Name": name, "Log": content, "LineCount": lineCount})
}

func clearLog(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/log/clear/")
	var apps map[string]models.Project
	_ = LoadJSON(config.ConfigApps, &apps)
	if app, ok := apps[name]; ok {
		_ = os.WriteFile(filepath.Join(app.Path, "run.log"), nil, 0644)
	}
	http.Redirect(w, r, "/log/"+name, 302)
}

func killSystemProc(w http.ResponseWriter, r *http.Request) {
	pid, _ := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/kill/"))
	if pid > 1 {
		if p, e := os.FindProcess(pid); e == nil {
			_ = p.Kill()
		}
	}
	http.Redirect(w, r, "/system", 302)
}
