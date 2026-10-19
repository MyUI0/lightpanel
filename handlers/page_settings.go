package handlers

import (
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
		"User":        usr,
		"FirstLogin":  isFirstLogin(r),
		"Username":    usr.Username,
		"BgUrl":       bgConfig.URL,
		"BgType":      bgConfig.Type,
		"LogoUrl":     getLogoURL(),
		"BarkEnabled": barkCfg.Enabled,
		"BarkDevice":  barkCfg.Device,
		"BarkGroup":   barkCfg.Group,
		"CustomPages": customPages,
		"Msg":         msg,
		"Err":         errMsg,
		"Sidebar":     template.HTML(sidebarHTML("/setting")),
		"Topbar":      template.HTML(topbarHTML("设置")),
	})
}

func saveAccount(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	newUsername := strings.TrimSpace(r.Form.Get("new_username"))
	oldPwd := r.Form.Get("old")
	newPwd := r.Form.Get("new")

	if oldPwd == "" {
		http.Redirect(w, r, "/setting?err=missing_pwd", 302)
		return
	}

	accountMu.Lock()
	defer accountMu.Unlock()

	var usr models.UserConfig
	if LoadJSON(config.ConfigUsr, &usr) != nil {
		http.Error(w, "无法读取账号", 500)
		return
	}

	if !passwordMatches(usr.Password, oldPwd) {
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

	if accountNeedsChange(usr) && newPwd == "" {
		http.Error(w, "请设置新密码", 400)
		return
	}
	if newPwd != "" {
		if len(newPwd) < 12 || len(newPwd) > 1024 || newPwd == oldPwd {
			http.Redirect(w, r, "/setting?err=password_weak", 302)
			return
		}
		encoded, err := passwordHash(newPwd)
		if err != nil {
			http.Error(w, "密码保存失败", 500)
			return
		}
		usr.Password = encoded
		usr.MustChangePassword = false
	}

	if err := WriteJSON(config.ConfigUsr, usr); err != nil {
		http.Error(w, "账号保存失败", 500)
		return
	}
	_ = os.Remove(filepath.Join(config.ConfigDir, "initial-password.txt"))
	logoutHandler(w, r)
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

	if _, err := archiveTarget(pagesDir, filename); err != nil {
		http.Error(w, "无效页面路径", 400)
		return
	}
	out, err := os.CreateTemp(pagesDir, ".page-*")
	if err != nil {
		http.Redirect(w, r, "/setting?err=create", 302)
		return
	}
	defer out.Close()
	defer os.Remove(out.Name())

	_, err = io.Copy(out, io.LimitReader(file, 10*1024*1024+1))
	if info, _ := out.Stat(); info != nil && info.Size() > 10*1024*1024 {
		http.Error(w, "页面超过10MB", 400)
		return
	}
	if err != nil {
		http.Redirect(w, r, "/setting?err=write", 302)
		return
	}

	if err = out.Close(); err != nil {
		http.Error(w, "保存失败", 500)
		return
	}
	if err = os.Rename(out.Name(), fpath); err != nil {
		http.Error(w, "保存失败", 500)
		return
	}

	http.Redirect(w, r, "/setting?ok=1", 302)
}

func deletePage(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/page/del/")
	pagesDir := config.DataDir + "/pages"

	if name == "" || strings.ContainsAny(name, "/\\") {
		http.Error(w, "无效页面名称", 400)
		return
	}
	fpath, err := archiveTarget(pagesDir, name)
	if err != nil {
		http.Error(w, "无效页面路径", 400)
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
	pid, err := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/kill/"))
	if err != nil || pid <= 1 {
		http.Error(w, "无效 PID", 400)
		return
	}
	appOpMu.Lock()
	defer appOpMu.Unlock()
	apps, err := loadAppsStrict()
	if err != nil {
		http.Error(w, "无法读取应用配置", 500)
		return
	}
	for name, app := range apps {
		checkAppStatus(&app)
		if app.PID == pid {
			killAppByName(name)
			app.PID = 0
			app.Status = "已停止"
			apps[name] = app
			if WriteJSON(config.ConfigApps, apps) != nil {
				http.Error(w, "保存停止状态失败", 500)
				return
			}
			http.Redirect(w, r, "/system", 303)
			return
		}
	}
	http.Error(w, "仅可停止身份已确认的受管进程", 403)
}
