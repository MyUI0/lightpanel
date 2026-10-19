package handlers

import (
	"html/template"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"lightpanel/config"
	"lightpanel/models"
)

func editAppHandler(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/edit/")
	if name == "" {
		http.Redirect(w, r, "/", 302)
		return
	}

	var csrfToken string
	cookie, _ := r.Cookie("lp_session")
	if cookie != nil {
		sessData := getSessionData(cookie.Value)
		if sessData != nil {
			csrfToken = sessData.CSRFToken
		}
	}

	if r.Method == "GET" {
		var apps map[string]models.Project
		_ = LoadJSON(config.ConfigApps, &apps)
		app, ok := apps[name]
		if !ok {
			http.Redirect(w, r, "/", 302)
			return
		}

		checkAppStatus(&app)

		msg := r.URL.Query().Get("ok")
		errMsg := r.URL.Query().Get("err")

		installNote := ""
		if b, err := os.ReadFile(filepath.Join(app.Path, "install_note.txt")); err == nil {
			installNote = strings.TrimSpace(string(b))
		}

		_ = htmlRender.ExecuteTemplate(w, "edit", map[string]any{
			"Name":        name,
			"Path":        app.Path,
			"Cmd":         app.Cmd,
			"SetupCmd":    app.SetupCmd,
			"WorkDir":     app.WorkDir,
			"SourceURL":   app.SourceURL,
			"URL":         app.URL,
			"Icon":        app.Icon,
			"Auto":        app.AutoStart,
			"Status":      app.Status,
			"Created":     app.Created,
			"Msg":         msg,
			"Err":         errMsg,
			"InstallNote": installNote,
			"CSRFToken":   csrfToken,
			"Sidebar":     template.HTML(sidebarHTML("/edit/")),
			"Topbar":      template.HTML(topbarHTML("编辑应用")),
		})
		return
	}

	if r.Method != "POST" {
		http.Redirect(w, r, "/", 302)
		return
	}

	appOpMu.Lock()
	defer appOpMu.Unlock()
	if err := r.ParseForm(); err != nil {
		http.Error(w, "无效表单", 400)
		return
	}
	apps, err := loadAppsStrict()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	original, ok := apps[name]
	if !ok {
		http.NotFound(w, r)
		return
	}
	app := original
	checkAppStatus(&app)
	if app.PID > 0 {
		http.Redirect(w, r, "/edit/"+url.PathEscape(name)+"?err=running", 303)
		return
	}
	newName, err := normalizedAppName(r.Form.Get("name"))
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	command, setup := strings.TrimSpace(r.Form.Get("cmd")), strings.TrimSpace(r.Form.Get("setup_cmd"))
	if command == "" || !validateCommand(command) || !validateCommand(setup) {
		http.Error(w, "命令无效", 400)
		return
	}
	oldPath, err := managedDirectory(original.Path)
	if err != nil {
		http.Error(w, "原应用路径无效", 400)
		return
	}
	target := strings.TrimSpace(r.Form.Get("path"))
	if target == "" {
		target = oldPath
	}
	if newName != name && pathKey(target) == pathKey(oldPath) {
		target = filepath.Join(config.AppsDir, newName)
	}
	target, err = managedDirectory(target)
	if err != nil {
		http.Error(w, "应用路径必须是应用目录的子目录", 400)
		return
	}
	for other, project := range apps {
		if other == name {
			continue
		}
		normalized, _ := normalizedAppName(other)
		if normalized == newName || pathsOverlap(project.Path, target) {
			http.Error(w, "应用名称或目录已占用", 409)
			return
		}
	}
	for reserved := range installReservations {
		if pathsOverlap(reserved, target) {
			http.Error(w, "目标目录已有安装任务", 409)
			return
		}
	}
	moved := newName != name && pathKey(oldPath) != pathKey(target)
	if !moved {
		if info, err := os.Stat(target); err != nil || !info.IsDir() {
			http.Error(w, "应用目录不存在", 400)
			return
		}
	}
	if moved {
		if _, err := os.Lstat(target); !os.IsNotExist(err) {
			http.Error(w, "重命名目标目录已存在", 409)
			return
		}
	}
	workDir := strings.TrimSpace(r.Form.Get("work_dir"))
	if workDir != "" {
		if !filepath.IsAbs(workDir) {
			workDir, err = archiveTarget(target, workDir)
			if err != nil {
				http.Error(w, "工作目录无效", 400)
				return
			}
		}
		if moved {
			if rel, e := filepath.Rel(oldPath, workDir); e == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
				workDir = filepath.Join(target, rel)
			}
		}
	}
	if moved {
		if err := os.Rename(oldPath, target); err != nil {
			http.Error(w, "重命名失败", 500)
			return
		}
	}
	app.Path = target
	app.Cmd = command
	app.SetupCmd = setup
	app.WorkDir = workDir
	app.AutoStart = r.Form.Get("auto") == "on"
	app.Icon = strings.TrimSpace(r.Form.Get("icon"))
	app.URL = strings.TrimSpace(r.Form.Get("url"))
	delete(apps, name)
	apps[newName] = app
	if err := WriteJSON(config.ConfigApps, apps); err != nil {
		if moved {
			if rollback := os.Rename(target, oldPath); rollback != nil {
				http.Error(w, "配置保存及目录回退失败，请检查目标目录", 500)
				return
			}
		}
		http.Error(w, "配置保存失败", 500)
		return
	}
	generations[name]++
	generations[newName]++
	delete(crashCounts, name)
	http.Redirect(w, r, "/edit/"+url.PathEscape(newName)+"?ok=1", 303)
}

func detectApp(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/detect/")
	if name == "" {
		http.Redirect(w, r, "/", 302)
		return
	}

	appOpMu.Lock()
	defer appOpMu.Unlock()

	var apps map[string]models.Project
	_ = LoadJSON(config.ConfigApps, &apps)
	app, ok := apps[name]
	if !ok {
		http.Redirect(w, r, "/", 302)
		return
	}

	scanTime := time.Now().Add(-60 * time.Second)
	result := combinedDetect(app.Path, scanTime)

	if result.WorkDir != "" {
		app.WorkDir = result.WorkDir
	}
	if result.Binary != "" {
		app.Cmd = result.Binary
	}
	apps[name] = app
	_ = WriteJSON(config.ConfigApps, apps)
	_ = os.WriteFile(filepath.Join(app.Path, "install_note.txt"), []byte(result.Note), 0644)

	http.Redirect(w, r, "/edit/"+name+"?ok=1", 302)
}
