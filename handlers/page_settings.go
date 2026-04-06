package handlers

import (
	"crypto/subtle"
	"html/template"
	"io"
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

	barkCfg := loadBarkConfig()

	customPages := getCustomPages()

	msg := r.URL.Query().Get("ok")
	errMsg := r.URL.Query().Get("err")

	_ = htmlRender.ExecuteTemplate(w, "setting", map[string]any{
		"User":         usr,
		"Username":     usr.Username,
		"BgUrl":        bgConfig.URL,
		"BgType":       bgConfig.Type,
		"LogoUrl":      getLogoURL(),
		"BarkEnabled":  barkCfg.Enabled,
		"BarkDevice":   barkCfg.Device,
		"BarkGroup":    barkCfg.Group,
		"CustomPages":  customPages,
		"Msg":          msg,
		"Err":          errMsg,
		"Sidebar":      template.HTML(sidebarHTML("/setting")),
		"Topbar":       template.HTML(topbarHTML("设置")),
	})
}

func saveAccount(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	newUsername := strings.TrimSpace(r.Form.Get("new_username"))
	oldPwd := strings.TrimSpace(r.Form.Get("old"))
	newPwd := strings.TrimSpace(r.Form.Get("new"))

	if oldPwd == "" {
		http.Redirect(w, r, "/setting?err=missing_pwd", 302)
		return
	}

	appOpMu.Lock()
	defer appOpMu.Unlock()

	var usr models.UserConfig
	_ = LoadJSON(config.ConfigUsr, &usr)

	if subtle.ConstantTimeCompare([]byte(hash(oldPwd)), []byte(usr.Password)) != 1 {
		http.Redirect(w, r, "/setting?err=password", 302)
		return
	}

	if newUsername != "" {
		if len(newUsername) < 3 {
			http.Redirect(w, r, "/setting?err=username_short", 302)
			return
		}
		usr.Username = newUsername
	}

	if newPwd != "" {
		if len(newPwd) < 6 {
			http.Redirect(w, r, "/setting?err=password_weak", 302)
			return
		}
		usr.Password = hash(newPwd)
		markPasswordChanged(w, r)
	}

	_ = WriteJSON(config.ConfigUsr, usr)
	http.Redirect(w, r, "/setting?ok=1", 302)
}

func saveBg(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()

	bgUrl := strings.TrimSpace(r.Form.Get("bg_url"))
	bgType := strings.TrimSpace(r.Form.Get("bg_type"))

	if bgType == "" {
		http.Redirect(w, r, "/setting?ok=1", 302)
		return
	}

	appOpMu.Lock()
	defer appOpMu.Unlock()

	bg := models.BgConfig{Type: bgType, URL: bgUrl}
	_ = WriteJSON(config.ConfigBg, bg)

	http.Redirect(w, r, "/setting?ok=1", 302)
}

func saveLogo(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()

	logoUrl := strings.TrimSpace(r.Form.Get("logo_url"))

	appOpMu.Lock()
	defer appOpMu.Unlock()

	type LogoConfig struct {
		URL string `json:"url"`
	}
	logo := LogoConfig{URL: logoUrl}
	_ = WriteJSON(config.ConfigDir+"/logo.json", logo)

	http.Redirect(w, r, "/setting?ok=1", 302)
}

func saveBark(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()

	enabled := r.Form.Get("bark_enabled") == "on"
	device := strings.TrimSpace(r.Form.Get("bark_device"))
	group := strings.TrimSpace(r.Form.Get("bark_group"))

	appOpMu.Lock()
	defer appOpMu.Unlock()

	cfg := BarkConfig{Enabled: enabled, Device: device, Group: group}
	_ = WriteJSON(config.ConfigDir+"/bark.json", cfg)

	http.Redirect(w, r, "/setting?ok=1", 302)
}

func uploadPage(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseMultipartForm(10 << 20)

	file, _, err := r.FormFile("page_file")
	if err != nil {
		http.Redirect(w, r, "/setting?err=upload", 302)
		return
	}
	defer file.Close()

	filename := strings.TrimSpace(r.Form.Get("filename"))
	if filename == "" {
		http.Redirect(w, r, "/setting?err=filename", 302)
		return
	}
	if !strings.HasSuffix(strings.ToLower(filename), ".html") {
		filename += ".html"
	}
	if strings.ContainsAny(filename, `/\`) || strings.Contains(filename, "..") {
		http.Redirect(w, r, "/setting?err=invalid", 302)
		return
	}

	pagesDir := config.DataDir + "/pages"
	_ = os.MkdirAll(pagesDir, 0755)
	fpath := filepath.Join(pagesDir, filename)

	out, err := os.Create(fpath)
	if err != nil {
		http.Redirect(w, r, "/setting?err=create", 302)
		return
	}
	defer out.Close()

	_, err = io.Copy(out, file)
	if err != nil {
		http.Redirect(w, r, "/setting?err=write", 302)
		return
	}

	http.Redirect(w, r, "/setting?ok=1", 302)
}

func deletePage(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/page/del/")
	pagesDir := config.DataDir + "/pages"
	fpath := filepath.Join(pagesDir, name)

	cleanPath := filepath.Clean(fpath)
	cleanDir := filepath.Clean(pagesDir)
	if !strings.HasPrefix(cleanPath, cleanDir) {
		http.Redirect(w, r, "/setting", 302)
		return
	}

	realPath, err := filepath.EvalSymlinks(fpath)
	if err != nil {
		http.Redirect(w, r, "/setting", 302)
		return
	}
	realDir, err := filepath.EvalSymlinks(pagesDir)
	if err != nil {
		http.Redirect(w, r, "/setting", 302)
		return
	}
	if !strings.HasPrefix(realPath, realDir) {
		http.Redirect(w, r, "/setting", 302)
		return
	}

	_ = os.Remove(fpath)
	http.Redirect(w, r, "/setting", 302)
}

func logPage(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/log/")
	var apps map[string]models.Project
	_ = LoadJSON(config.ConfigApps, &apps)

	if name == "" {
		appList := make([]string, 0, len(apps))
		for n := range apps {
			appList = append(appList, n)
		}
		_ = htmlRender.ExecuteTemplate(w, "log_list", map[string]any{
			"Apps":    appList,
			"Sidebar": template.HTML(sidebarHTML("/log/")),
			"Topbar":  template.HTML(topbarHTML("运行日志")),
		})
		return
	}

	app, ok := apps[name]
	if !ok {
		http.Redirect(w, r, "/log/", 302)
		return
	}
	logFile := filepath.Join(app.Path, "run.log")
	content := ""
	lineCount := 0
	if info, e := os.Stat(logFile); e == nil {
		f, e := os.Open(logFile)
		if e == nil {
			size := info.Size()
			readSize := size
			if readSize > int64(config.MaxLogLen) {
				readSize = int64(config.MaxLogLen)
			}
			if readSize < 1 {
				readSize = 1
			}
			offset := size - readSize
			if offset < 0 {
				offset = 0
			}
			f.Seek(offset, 0)
			buf := make([]byte, readSize)
			n, _ := f.Read(buf)
			_ = f.Close()
			if n > 0 {
				content = string(buf[:n])
				if idx := strings.Index(content, "\n"); idx >= 0 && offset > 0 {
					content = content[idx+1:]
				}
			}
		}
	}
	lineCount = strings.Count(content, "\n")
	_ = htmlRender.ExecuteTemplate(w, "log", map[string]any{"Name": name, "Log": content, "LineCount": lineCount, "Sidebar": template.HTML(sidebarHTML("/log/")), "Topbar": template.HTML(topbarHTML("运行日志"))})
}

func clearLog(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/log/clear/")
	var apps map[string]models.Project
	_ = LoadJSON(config.ConfigApps, &apps)
	if app, ok := apps[name]; ok {
		_ = os.WriteFile(filepath.Join(app.Path, "run.log"), []byte{}, 0644)
	}
	http.Redirect(w, r, "/log/"+name, 302)
}

func killSystemProc(w http.ResponseWriter, r *http.Request) {
	pid, _ := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/kill/"))
	if pid <= 1 || pid >= 100000 {
		http.Redirect(w, r, "/system", 302)
		return
	}

	if protectedPids[pid] {
		http.Redirect(w, r, "/system", 302)
		return
	}

	var apps map[string]models.Project
	_ = LoadJSON(config.ConfigApps, &apps)
	allowed := false
	for _, app := range apps {
		pidFile := filepath.Join(app.Path, "pid.pid")
		if b, e := os.ReadFile(pidFile); e == nil {
			appPid, err := strconv.Atoi(strings.TrimSpace(string(b)))
			if err == nil && appPid == pid {
				allowed = true
				break
			}
		}
	}
	if !allowed {
		http.Redirect(w, r, "/system", 302)
		return
	}

	if p, e := os.FindProcess(pid); e == nil {
		_ = p.Kill()
	}
	http.Redirect(w, r, "/system", 302)
}

var protectedPids = map[int]bool{
	1: true, 2: true, 3: true, 4: true, 5: true, 6: true, 7: true, 8: true, 9: true,
	10: true, 11: true, 12: true, 13: true, 14: true, 15: true, 16: true, 17: true, 18: true,
	19: true, 20: true, 21: true, 22: true, 23: true, 24: true, 25: true, 26: true, 27: true, 28: true,
	29: true, 30: true,
}
