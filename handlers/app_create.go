package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
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

	_ = r.ParseForm()

	name := strings.TrimSpace(r.FormValue("name"))
	down := strings.TrimSpace(r.FormValue("url"))
	cmd := strings.TrimSpace(r.FormValue("cmd"))
	setupCmd := strings.TrimSpace(r.FormValue("setup_cmd"))
	autoExtract := r.FormValue("auto_extract") == "on"
	makeExec := r.FormValue("make_exec") == "on"

	if name == "" {
		json.NewEncoder(w).Encode(map[string]any{"error": "应用名称不能为空"})
		return
	}
	if down == "" {
		json.NewEncoder(w).Encode(map[string]any{"error": "下载地址不能为空"})
		return
	}
	if !validateCommand(cmd) {
		json.NewEncoder(w).Encode(map[string]any{"error": "启动命令包含非法字符"})
		return
	}
	if !validateCommand(setupCmd) {
		json.NewEncoder(w).Encode(map[string]any{"error": "首次运行命令包含非法字符"})
		return
	}
	if strings.ContainsAny(name, "./\\") || strings.ContainsAny(name, ";&|'") || strings.HasPrefix(name, ".") || strings.Contains(name, "..") || strings.ContainsAny(name, "?*:#") {
		json.NewEncoder(w).Encode(map[string]any{"error": "应用名称包含非法字符"})
		return
	}

	appOpMu.Lock()
	var apps map[string]models.Project
	_ = LoadJSON(config.ConfigApps, &apps)
	if _, exists := apps[name]; exists {
		appOpMu.Unlock()
		json.NewEncoder(w).Encode(map[string]any{"error": "应用已存在"})
		return
	}
	appOpMu.Unlock()

	taskID := "create_" + strconv.FormatInt(time.Now().UnixNano(), 36)
	task := &CreateTask{
		ID:        taskID,
		Name:      name,
		Progress:  0,
		Status:    "creating",
		Message:   "准备创建...",
		CreatedAt: time.Now(),
	}
	createTasks.Store(taskID, task)

	go func() {
		defer func() { recover() }()
		dirName := strings.ReplaceAll(name, " ", "-")
		sandbox := filepath.Join(config.AppsDir, dirName)
		if err := os.MkdirAll(sandbox, 0755); err != nil {
			task.SetStatus("error", "创建目录失败")
			return
		}

		if down != "" {
			if !strings.HasPrefix(down, "http://") && !strings.HasPrefix(down, "https://") {
				task.SetStatus("error", "URL 格式错误")
				_ = os.RemoveAll(sandbox)
				return
			}
			if isPrivateURL(down) {
				task.SetStatus("error", "不允许访问内网地址")
				_ = os.RemoveAll(sandbox)
				return
			}

			dlID := "dl_" + strconv.FormatInt(time.Now().UnixNano(), 36)
			dlTask := &models.DownloadTask{
				ID:     dlID,
				Name:   name,
				URL:    down,
				Cmd:    cmd,
				Status: "downloading",
				Last:   time.Now(),
			}
			downloadTasks.Store(dlID, dlTask)
			task.SetProgress(10, "正在下载...")

			client := &http.Client{Timeout: 120 * time.Second}
			resp, e := client.Get(down)
			if e != nil {
				dlTask.SetStatus("error")
				dlTask.Last = time.Now()
				task.SetStatus("error", "下载失败: "+e.Error())
				_ = os.RemoveAll(sandbox)
				return
			}
			defer resp.Body.Close()
			if resp.StatusCode != 200 {
				dlTask.SetStatus("error")
				dlTask.Last = time.Now()
				task.SetStatus("error", "HTTP "+fmt.Sprint(resp.StatusCode))
				_ = os.RemoveAll(sandbox)
				return
			}

			dlTask.UpdateSize(resp.ContentLength)
			fname := filepath.Base(down)
			if fname == "" || fname == "." || fname == "/" {
				fname = "download"
			}
			fpath := filepath.Join(sandbox, fname)
			f, err := os.Create(fpath)
			if err != nil {
				dlTask.SetStatus("error")
				dlTask.Last = time.Now()
				task.SetStatus("error", "创建文件失败")
				_ = os.RemoveAll(sandbox)
				return
			}

			buf := make([]byte, 32*1024)
			var totalWritten int64
			var dlSize = dlTask.GetSize()
			for {
				n, readErr := resp.Body.Read(buf)
				if n > 0 {
					if _, writeErr := f.Write(buf[:n]); writeErr != nil {
						_ = f.Close()
						dlTask.SetStatus("error")
						dlTask.Last = time.Now()
						task.SetStatus("error", "写入失败")
						_ = os.Remove(fpath)
						_ = os.RemoveAll(sandbox)
						return
					}
					totalWritten += int64(n)
					if dlSize > 0 {
						prog := int(float64(totalWritten) / float64(dlSize) * 90)
						dlTask.UpdateProgress(totalWritten, dlSize, prog)
						_, _, msg := task.Get()
						task.SetProgress(10+prog/10, msg)
					}
				}
				if readErr != nil {
					if !errors.Is(readErr, io.EOF) {
						_ = f.Close()
						dlTask.SetStatus("error")
						dlTask.Last = time.Now()
						task.SetStatus("error", "读取失败")
						_ = os.Remove(fpath)
						_ = os.RemoveAll(sandbox)
						return
					}
					break
				}
				if totalWritten >= int64(config.MaxDownBytes) {
					break
				}
			}
			_ = f.Close()

			dlTask.SetStatus("completed")
			dlTask.UpdateProgress(dlTask.GetDownloaded(), dlSize, 100)
			task.SetProgress(50, "")

		// autoExtract 仅在有下载文件时生效（嵌套在 if down != "" 内部）
		// 手动添加应用时即使设置了 auto_extract 也不会执行解压逻辑
		if autoExtract {
			task.SetProgress(50, "正在解压...")
			if extractArchive(fpath, sandbox) {
				_ = os.Remove(fpath)
				if cmd == "" {
					if bin := findBinaryInDir(sandbox); bin != "" {
						cmd = bin
					}
				}
			} else if makeExec && strings.HasSuffix(fname, ".sh") {
				task.SetProgress(50, "设置权限...")
				_ = os.Chmod(fpath, 0755)
				if cmd == "" {
					cmd = "./" + fname
				}
			}
		}
		}

		task.SetProgress(80, "保存配置...")

		appOpMu.Lock()
		var finalApps map[string]models.Project
		_ = LoadJSON(config.ConfigApps, &finalApps)
		finalApps[name] = models.Project{
			Path:      sandbox,
			Cmd:       cmd,
			SetupCmd:  setupCmd,
			WorkDir:   "",
			SourceURL: down,
			Created:   time.Now().Format("2006-01-02 15:04"),
		}
		_ = WriteJSON(config.ConfigApps, finalApps)
		appOpMu.Unlock()

		task.SetStatus("completed", "创建完成")
		task.SetProgress(100, "")

		time.AfterFunc(5*time.Minute, func() {
			createTasks.Delete(taskID)
		})
	}()

	json.NewEncoder(w).Encode(map[string]any{
		"ok":       true,
		"task":     taskID,
		"redirect": "/",
	})
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

	cleanPath := filepath.Clean(path)
	cleanAppsDir := filepath.Clean(config.AppsDir)
	realPath, err := filepath.EvalSymlinks(cleanPath)
	if err != nil {
		realPath = cleanPath
	}
	realAppsDir, _ := filepath.EvalSymlinks(cleanAppsDir)
	if realAppsDir == "" {
		realAppsDir = cleanAppsDir
	}
	if !strings.HasPrefix(realPath+string(os.PathSeparator), realAppsDir+string(os.PathSeparator)) && realPath != realAppsDir {
		json.NewEncoder(w).Encode(map[string]any{"error": "路径必须在应用目录内"})
		return
	}

	if strings.ContainsAny(name, "./\\") || strings.ContainsAny(name, ";&|'") || strings.HasPrefix(name, ".") || strings.Contains(name, "..") || strings.ContainsAny(name, "?*:#") {
		json.NewEncoder(w).Encode(map[string]any{"error": "应用名称包含非法字符"})
		return
	}

	appOpMu.Lock()
	var apps map[string]models.Project
	_ = LoadJSON(config.ConfigApps, &apps)
	if _, exists := apps[name]; exists {
		appOpMu.Unlock()
		json.NewEncoder(w).Encode(map[string]any{"error": "应用已存在"})
		return
	}

	apps[name] = models.Project{
		Path:      path,
		Cmd:       cmd,
		WorkDir:   workDir,
		Icon:     icon,
		URL:       url,
		AutoStart: auto,
		Created:   time.Now().Format("2006-01-02 15:04"),
	}
	_ = WriteJSON(config.ConfigApps, apps)
	appOpMu.Unlock()

	json.NewEncoder(w).Encode(map[string]any{"ok": true})
}
