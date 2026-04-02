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

func startApp(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/start/")
	var apps map[string]models.Project
	_ = LoadJSON(config.ConfigApps, &apps)
	app, ok := apps[name]
	if !ok {
		http.Redirect(w, r, "/", 302)
		return
	}

	if launchApp(name, app) {
		app.AutoStart = true
		apps[name] = app
		_ = WriteJSON(config.ConfigApps, apps)
	}
	http.Redirect(w, r, "/", 302)
}

func stopApp(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/stop/")
	var apps map[string]models.Project
	_ = LoadJSON(config.ConfigApps, &apps)
	app, ok := apps[name]
	if !ok {
		http.Redirect(w, r, "/", 302)
		return
	}

	pidFile := filepath.Join(app.Path, "pid.pid")
	if b, e := os.ReadFile(pidFile); e == nil {
		if pid, err := strconv.Atoi(strings.TrimSpace(string(b))); err == nil && pid > 0 {
			killByPid(pid)
		}
		_ = os.Remove(pidFile)
	}
	http.Redirect(w, r, "/", 302)
}

func restartApp(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/restart/")
	var apps map[string]models.Project
	_ = LoadJSON(config.ConfigApps, &apps)
	app, ok := apps[name]
	if !ok {
		http.Redirect(w, r, "/", 302)
		return
	}

	pidFile := filepath.Join(app.Path, "pid.pid")
	if b, e := os.ReadFile(pidFile); e == nil {
		if pid, err := strconv.Atoi(strings.TrimSpace(string(b))); err == nil && pid > 0 {
			killByPid(pid)
			for i := 0; i < 20; i++ {
				if p, e := os.FindProcess(pid); e != nil || p.Signal(syscall.Signal(0)) != nil {
					break
				}
				time.Sleep(100 * time.Millisecond)
			}
		}
		_ = os.Remove(pidFile)
	}

	if launchApp(name, app) {
		app.AutoStart = true
		apps[name] = app
		_ = WriteJSON(config.ConfigApps, apps)
	}
	http.Redirect(w, r, "/", 302)
}

func deleteApp(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/delete/")
	var apps map[string]models.Project
	_ = LoadJSON(config.ConfigApps, &apps)
	app, ok := apps[name]
	if !ok {
		http.Redirect(w, r, "/", 302)
		return
	}

	pidFile := filepath.Join(app.Path, "pid.pid")
	if b, e := os.ReadFile(pidFile); e == nil {
		if pid, err := strconv.Atoi(strings.TrimSpace(string(b))); err == nil && pid > 0 {
			killByPid(pid)
			for i := 0; i < 20; i++ {
				if p, e := os.FindProcess(pid); e != nil || p.Signal(syscall.Signal(0)) != nil {
					break
				}
				time.Sleep(100 * time.Millisecond)
			}
		}
		_ = os.Remove(pidFile)
	}

	delete(apps, name)
	_ = WriteJSON(config.ConfigApps, apps)

	_ = os.RemoveAll(app.Path)

	http.Redirect(w, r, "/", 302)
}

func toggleAutoStart(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/toggle-auto/")
	var apps map[string]models.Project
	_ = LoadJSON(config.ConfigApps, &apps)
	app, ok := apps[name]
	if !ok {
		http.Redirect(w, r, "/", 302)
		return
	}

	app.AutoStart = !app.AutoStart
	apps[name] = app
	_ = WriteJSON(config.ConfigApps, apps)

	http.Redirect(w, r, "/", 302)
}
