package handlers

import (
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
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
		"Auto":        app.AutoStart,
		"Status":      app.Status,
		"Created":     app.Created,
		"Msg":         msg,
		"Err":         errMsg,
		"InstallNote": installNote,
	})
		return
	}

	appOpMu.Lock()
	defer appOpMu.Unlock()

	_ = r.ParseForm()

	newName := strings.TrimSpace(r.Form.Get("name"))
	newPath := strings.TrimSpace(r.Form.Get("path"))
	newCmd := strings.TrimSpace(r.Form.Get("cmd"))
	newSetupCmd := strings.TrimSpace(r.Form.Get("setup_cmd"))
	autoStr := strings.TrimSpace(r.Form.Get("auto"))
	newWorkDir := strings.TrimSpace(r.Form.Get("work_dir"))
	newURL := strings.TrimSpace(r.Form.Get("url"))

	if newName == "" || newCmd == "" || strings.ContainsAny(newName, `./\ `) || strings.HasPrefix(newName, ".") || strings.Contains(newName, "..") {
		http.Redirect(w, r, "/edit/"+name+"?err=invalid", 302)
		return
	}
	if !validateCommand(newCmd) || !validateCommand(newSetupCmd) {
		http.Redirect(w, r, "/edit/"+name+"?err=invalid_cmd", 302)
		return
	}

	var apps map[string]models.Project
	_ = LoadJSON(config.ConfigApps, &apps)

	app, ok := apps[name]
	if !ok {
		http.Redirect(w, r, "/", 302)
		return
	}

	running := false
	pidFile := filepath.Join(app.Path, "pid.pid")
	if b, e := os.ReadFile(pidFile); e == nil {
		if pid, err := strconv.Atoi(strings.TrimSpace(string(b))); err == nil && pid > 0 {
			if p, e := os.FindProcess(pid); e == nil && p.Signal(syscall.Signal(0)) == nil {
				running = true
			}
		}
	}

	if running {
		http.Redirect(w, r, "/edit/"+name+"?err=running", 302)
		return
	}

	oldPath := app.Path
	autoStart := autoStr == "on"
	isRename := newName != name

	if isRename {
		if _, exists := apps[newName]; exists {
			http.Redirect(w, r, "/edit/"+name+"?err=exists", 302)
			return
		}
	}

	app.Cmd = newCmd
	app.SetupCmd = newSetupCmd
	app.AutoStart = autoStart
	app.WorkDir = newWorkDir
	app.URL = newURL

	if newPath != "" {
		app.Path = newPath
	} else if isRename {
		app.Path = filepath.Join(config.AppsDir, newName)
	}

	apps[newName] = app
	if isRename {
		delete(apps, name)
	}

	if err := WriteJSON(config.ConfigApps, apps); err != nil {
		http.Redirect(w, r, "/edit/"+name+"?err=save", 302)
		return
	}

	if isRename {
		newDirPath := app.Path
		if app.Path == oldPath {
			newDirPath = filepath.Join(config.AppsDir, newName)
		}
		if err := os.Rename(oldPath, newDirPath); err != nil {
			app.Path = oldPath
			apps[name] = app
			delete(apps, newName)
			_ = WriteJSON(config.ConfigApps, apps)
			http.Redirect(w, r, "/edit/"+name+"?err=rename", 302)
			return
		}
		app.Path = newDirPath
		apps[newName] = app
		_ = WriteJSON(config.ConfigApps, apps)
	}

	http.Redirect(w, r, "/edit/"+newName+"?ok=1", 302)
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
