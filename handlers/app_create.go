package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"lightpanel/config"
	"lightpanel/models"
)

var createTasks sync.Map

type CreateTask struct {
	mu        sync.Mutex
	ID        string
	Name      string
	Progress  int
	Status    string
	Message   string
	CreatedAt time.Time
}

func (t *CreateTask) SetProgress(progress int, message string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.Progress = progress
	t.Message = message
}

func (t *CreateTask) SetStatus(status string, message string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.Status = status
	if message != "" {
		t.Message = message
	}
}

func (t *CreateTask) Get() (Progress int, Status string, Message string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.Progress, t.Status, t.Message
}

func apiCreateProgress(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	taskID := strings.TrimPrefix(r.URL.Path, "/create/progress/")
	val, ok := createTasks.Load(taskID)
	if !ok {
		json.NewEncoder(w).Encode(map[string]any{"status": "not_found"})
		return
	}
	task := val.(*CreateTask)
	progress, status, message := task.Get()
	json.NewEncoder(w).Encode(map[string]any{
		"status":   status,
		"progress": progress,
		"message":  message,
	})
}

func createApp(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	app := models.StoreApp{Name: r.FormValue("name"), URL: strings.TrimSpace(r.FormValue("url")), Cmd: strings.TrimSpace(r.FormValue("cmd")), SetupCmd: strings.TrimSpace(r.FormValue("setup_cmd")), AutoExtract: r.FormValue("auto_extract") == "on", MakeExec: r.FormValue("make_exec") == "on"}
	if app.URL == "" {
		json.NewEncoder(w).Encode(map[string]any{"error": "下载地址不能为空"})
		return
	}
	task := &CreateTask{ID: "create_" + generateSessionToken(), Name: app.Name, Status: "creating", CreatedAt: time.Now()}
	createTasks.Store(task.ID, task)
	if _, err := submitInstall(installPlan{App: app}, task); err != nil {
		createTasks.Delete(task.ID)
		json.NewEncoder(w).Encode(map[string]any{"error": err.Error()})
		return
	}
	json.NewEncoder(w).Encode(map[string]any{"ok": true, "task": task.ID, "redirect": "/"})
}

func createManualApp(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	_ = r.ParseForm()

	name := strings.TrimSpace(r.Form.Get("name"))
	path := strings.TrimSpace(r.Form.Get("path"))
	cmd := strings.TrimSpace(r.Form.Get("cmd"))
	workDir := strings.TrimSpace(r.Form.Get("work_dir"))
	icon := strings.TrimSpace(r.Form.Get("icon"))
	url := strings.TrimSpace(r.Form.Get("url"))
	auto := r.Form.Get("auto") == "on"

	if name == "" {
		json.NewEncoder(w).Encode(map[string]any{"error": "应用名称不能为空"})
		return
	}
	if path == "" {
		json.NewEncoder(w).Encode(map[string]any{"error": "应用目录路径不能为空"})
		return
	}
	if cmd == "" {
		json.NewEncoder(w).Encode(map[string]any{"error": "启动命令不能为空"})
		return
	}
	if !validateCommand(cmd) {
		json.NewEncoder(w).Encode(map[string]any{"error": "启动命令包含非法字符"})
		return
	}

	canonical, err := managedDirectory(path)
	if err != nil {
		http.Error(w, "路径必须是应用目录的有效子目录", 400)
		return
	}
	path = canonical
	if info, err := os.Stat(path); err != nil || !info.IsDir() {
		http.Error(w, "应用目录不存在", 400)
		return
	}
	if workDir != "" && !filepath.IsAbs(workDir) {
		workDir, err = archiveTarget(path, workDir)
		if err != nil {
			http.Error(w, "工作目录无效", 400)
			return
		}
	}

	if strings.ContainsAny(name, "./\\") || strings.ContainsAny(name, ";&|'") || strings.HasPrefix(name, ".") || strings.Contains(name, "..") || strings.ContainsAny(name, "?*:#") {
		json.NewEncoder(w).Encode(map[string]any{"error": "应用名称包含非法字符"})
		return
	}

	appOpMu.Lock()
	defer appOpMu.Unlock()
	apps, loadErr := loadAppsStrict()
	if loadErr != nil {
		http.Error(w, loadErr.Error(), 500)
		return
	}
	if _, exists := apps[name]; exists {
		json.NewEncoder(w).Encode(map[string]any{"error": "应用已存在"})
		return
	}

	for n, p := range apps {
		normalized, _ := normalizedAppName(n)
		wanted, _ := normalizedAppName(name)
		if normalized == wanted || pathsOverlap(p.Path, path) {
			http.Error(w, "应用名称或目录已被占用", 409)
			return
		}
	}
	for reserved := range installReservations {
		if pathsOverlap(reserved, path) {
			http.Error(w, "目录已有安装任务", 409)
			return
		}
	}
	apps[name] = models.Project{
		Path:      path,
		Cmd:       cmd,
		WorkDir:   workDir,
		Icon:      icon,
		URL:       url,
		AutoStart: auto,
		Created:   time.Now().Format("2006-01-02 15:04"),
	}
	if err := WriteJSON(config.ConfigApps, apps); err != nil {
		http.Error(w, "保存应用失败", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]any{"ok": true})
}
