package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"lightpanel/config"
	"lightpanel/models"
)

var appOpMu sync.Mutex

func killAppByName(name string) {
	generations[name]++
	delete(crashCounts, name)
	var apps map[string]models.Project
	_ = LoadJSON(config.ConfigApps, &apps)
	app, ok := apps[name]
	if !ok {
		return
	}
	checkAppStatus(&app)
	if app.PID > 0 {
		killAppProcess(app.PID)
	}
	_ = os.Remove(filepath.Join(app.Path, "pid.pid"))
}

func controlReply(w http.ResponseWriter, r *http.Request, err error) {
	if strings.Contains(r.Header.Get("Accept"), "application/json") {
		w.Header().Set("Content-Type", "application/json")
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]any{"ok": false, "error": err.Error()})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]bool{"ok": true})
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func startApp(w http.ResponseWriter, r *http.Request)   { controlApp(w, r, "start") }
func stopApp(w http.ResponseWriter, r *http.Request)    { controlApp(w, r, "stop") }
func restartApp(w http.ResponseWriter, r *http.Request) { controlApp(w, r, "restart") }
func controlApp(w http.ResponseWriter, r *http.Request, action string) {
	name := strings.TrimPrefix(r.URL.Path, "/"+action+"/")
	appOpMu.Lock()
	defer appOpMu.Unlock()
	var apps map[string]models.Project
	if err := LoadJSON(config.ConfigApps, &apps); err != nil {
		controlReply(w, r, err)
		return
	}
	app, ok := apps[name]
	if !ok {
		controlReply(w, r, fmt.Errorf("应用不存在"))
		return
	}
	if action == "stop" || action == "restart" {
		killAppByName(name)
	}
	if action == "stop" {
		app.Status, app.PID = "已停止", 0
		go RunHook("app_stop", name)
	} else {
		delete(crashCounts, name)
		checkAppStatus(&app)
		if app.PID == 0 && !launchApp(name, &app) {
			controlReply(w, r, fmt.Errorf("启动失败，请检查应用日志与启动命令"))
			return
		}
	}
	apps[name] = app
	controlReply(w, r, WriteJSON(config.ConfigApps, apps))
}

func deleteApp(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/delete/")
	appOpMu.Lock()
	defer appOpMu.Unlock()
	var apps map[string]models.Project
	if err := LoadJSON(config.ConfigApps, &apps); err != nil {
		controlReply(w, r, err)
		return
	}
	app, ok := apps[name]
	if !ok {
		controlReply(w, r, fmt.Errorf("应用不存在"))
		return
	}
	base, err := filepath.Abs(config.AppsDir)
	if err != nil {
		controlReply(w, r, err)
		return
	}
	target, err := filepath.Abs(app.Path)
	if err != nil {
		controlReply(w, r, err)
		return
	}
	rel, err := filepath.Rel(base, target)
	if err != nil || rel == "." || rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		controlReply(w, r, fmt.Errorf("应用目录超出允许删除范围"))
		return
	}
	if _, err := archiveTarget(base, rel); err != nil {
		controlReply(w, r, err)
		return
	}
	killAppByName(name)
	delete(apps, name)
	if err := WriteJSON(config.ConfigApps, apps); err != nil {
		controlReply(w, r, err)
		return
	}
	if err := os.RemoveAll(target); err != nil {
		apps[name] = app
		if rollbackErr := WriteJSON(config.ConfigApps, apps); rollbackErr != nil {
			controlReply(w, r, fmt.Errorf("删除目录失败，恢复配置也失败：%v", rollbackErr))
			return
		}
		controlReply(w, r, fmt.Errorf("删除目录失败：%v", err))
		return
	}
	controlReply(w, r, nil)
}

func toggleAutoStart(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/toggle-auto/")
	appOpMu.Lock()
	defer appOpMu.Unlock()
	var apps map[string]models.Project
	if err := LoadJSON(config.ConfigApps, &apps); err != nil || apps == nil {
		controlReply(w, r, fmt.Errorf("读取应用配置失败"))
		return
	}
	app, ok := apps[name]
	if !ok {
		controlReply(w, r, fmt.Errorf("应用不存在"))
		return
	}

	app.AutoStart = !app.AutoStart
	apps[name] = app
	if err := WriteJSON(config.ConfigApps, apps); err != nil {
		controlReply(w, r, err)
		return
	}

	controlReply(w, r, nil)
}
