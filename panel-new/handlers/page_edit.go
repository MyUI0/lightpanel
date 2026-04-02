package handlers

import (
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

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

		_ = htmlRender.ExecuteTemplate(w, "edit", map[string]any{
			"Name":    name,
			"Path":    app.Path,
			"Cmd":     app.Cmd,
			"Auto":    app.AutoStart,
			"Status":  app.Status,
			"Created": app.Created,
			"Msg":     msg,
			"Err":     errMsg,
		})
		return
	}

	_ = r.ParseForm()

	newName := strings.TrimSpace(r.Form.Get("name"))
	newPath := strings.TrimSpace(r.Form.Get("path"))
	newCmd := strings.TrimSpace(r.Form.Get("cmd"))
	autoStr := strings.TrimSpace(r.Form.Get("auto"))

	if newName == "" || newCmd == "" || strings.ContainsAny(newName, `./\ `) {
		http.Redirect(w, r, "/edit/"+name+"?err=invalid", 302)
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

	if newName != name {
		if _, exists := apps[newName]; exists {
			http.Redirect(w, r, "/edit/"+name+"?err=exists", 302)
			return
		}

		newSandbox := filepath.Join(config.AppsDir, newName)
		app.Path = newSandbox
	}

	if newPath != "" && newPath != app.Path {
		app.Path = newPath
	}

	app.Cmd = newCmd
	app.AutoStart = autoStart

	apps[newName] = app
	if newName != name {
		delete(apps, name)
	}

	if err := WriteJSON(config.ConfigApps, apps); err != nil {
		http.Redirect(w, r, "/edit/"+name+"?err=save", 302)
		return
	}

	if newName != name {
		newSandbox := filepath.Join(config.AppsDir, newName)
		if err := os.Rename(oldPath, newSandbox); err != nil {
			app.Path = oldPath
			apps[name] = app
			delete(apps, newName)
			_ = WriteJSON(config.ConfigApps, apps)
			http.Redirect(w, r, "/edit/"+name+"?err=rename", 302)
			return
		}
	}

	http.Redirect(w, r, "/edit/"+newName+"?ok=1", 302)
}
