package handlers

import (
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"lightpanel/config"
	"lightpanel/models"
)

var appOpMu sync.Mutex

func killAppByName(name string) {
	var apps map[string]models.Project
	_ = LoadJSON(config.ConfigApps, &apps)
	app, ok := apps[name]
	if !ok {
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
}

func startApp(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/start/")
	appOpMu.Lock()
	defer appOpMu.Unlock()
	var apps map[string]models.Project
	_ = LoadJSON(config.ConfigApps, &apps)
	app, ok := apps[name]
	if !ok {
		http.Redirect(w, r, "/", 302)
		return
	}

	if launchApp(name, &app) {
		apps[name] = app
		_ = WriteJSON(config.ConfigApps, apps)
	}
	http.Redirect(w, r, "/", 302)
}

func stopApp(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/stop/")
	appOpMu.Lock()
	defer appOpMu.Unlock()
	killAppByName(name)
	go RunHook("app_stop", name)
	http.Redirect(w, r, "/", 302)
}

func restartApp(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/restart/")
	appOpMu.Lock()
	defer appOpMu.Unlock()
	var apps map[string]models.Project
	_ = LoadJSON(config.ConfigApps, &apps)
	app, ok := apps[name]
	if !ok {
		http.Redirect(w, r, "/", 302)
		return
	}

	killAppByName(name)

	if launchApp(name, &app) {
		apps[name] = app
		_ = WriteJSON(config.ConfigApps, apps)
	}
	http.Redirect(w, r, "/", 302)
}

func deleteApp(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/delete/")
	appOpMu.Lock()
	defer appOpMu.Unlock()
	var apps map[string]models.Project
	_ = LoadJSON(config.ConfigApps, &apps)
	app, ok := apps[name]
	if !ok {
		http.Redirect(w, r, "/", 302)
		return
	}

	killAppByName(name)

	delete(apps, name)
	_ = WriteJSON(config.ConfigApps, apps)

	cleanPath := filepath.Clean(app.Path)
	cleanAppsDir := filepath.Clean(config.AppsDir)
	realPath, err := filepath.EvalSymlinks(cleanPath)
	if err != nil {
		realPath = cleanPath
	}
	realAppsDir, err := filepath.EvalSymlinks(cleanAppsDir)
	if err != nil {
		realAppsDir = cleanAppsDir
	}
	if strings.HasPrefix(realPath+string(os.PathSeparator), realAppsDir+string(os.PathSeparator)) || realPath == realAppsDir {
		if realPath != realAppsDir {
			_ = os.RemoveAll(cleanPath)
		}
	}

	http.Redirect(w, r, "/", 302)
}

func toggleAutoStart(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/toggle-auto/")
	appOpMu.Lock()
	defer appOpMu.Unlock()
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
