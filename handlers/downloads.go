package handlers

import (
	"encoding/json"
	"errors"
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

var (
	downloadTasks sync.Map
)

func init() {
	ticker := time.NewTicker(10 * time.Minute)
	go func() {
		defer func() { recover() }()
		defer ticker.Stop()
		for range ticker.C {
			downloadTasks.Range(func(key, value any) bool {
				if t, ok := value.(*models.DownloadTask); ok {
					status := t.GetStatus()
					if status == "completed" || status == "error" {
						if time.Since(t.Last) > 30*time.Minute {
							downloadTasks.Delete(key)
						}
					} else if status == "downloading" && time.Since(t.Last) > 60*time.Minute {
						t.SetStatus("error")
						downloadTasks.Delete(key)
					}
				}
				return true
			})
		}
	}()
}

func getDownloadTask(id string) (*models.DownloadTask, bool) {
	val, ok := downloadTasks.Load(id)
	if !ok {
		return nil, false
	}
	t, ok := val.(*models.DownloadTask)
	return t, ok
}

func snapshotTasks() []models.DownloadTask {
	var tasks []models.DownloadTask
	downloadTasks.Range(func(_, value any) bool {
		if t, ok := value.(*models.DownloadTask); ok {
			tasks = append(tasks, t.Snapshot())
		}
		return true
	})
	return tasks
}

func downloadPage(w http.ResponseWriter, r *http.Request) {
	tasks := snapshotTasks()
	_ = htmlRender.ExecuteTemplate(w, "downloads", tasks)
}

func apiDownloads(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	tasks := snapshotTasks()
	_ = json.NewEncoder(w).Encode(tasks)
}

func apiDownloadAction(w http.ResponseWriter, r *http.Request) {
	action := strings.TrimPrefix(r.URL.Path, "/dl/action/")
	parts := strings.SplitN(action, "/", 2)
	if len(parts) != 2 {
		w.WriteHeader(400)
		return
	}
	id, act := parts[0], parts[1]

	task, ok := getDownloadTask(id)
	if !ok {
		w.WriteHeader(404)
		return
	}

	switch act {
	case "pause":
		if task.GetStatus() == "downloading" {
			task.SetStatus("paused")
		}
	case "resume":
		if task.GetStatus() == "paused" {
			task.SetStatus("downloading")
			go resumeDownload(id)
		}
	case "delete":
		downloadTasks.Delete(id)
		if s := task.GetStatus(); s == "downloading" || s == "paused" {
			fpath := filepath.Join(config.AppsDir, task.Name, filepath.Base(task.URL))
			_ = os.Remove(fpath)
		}
	case "install":
		if task.GetStatus() == "completed" {
			installDownloadedApp(id)
			downloadTasks.Delete(id)
		}
	}

	w.WriteHeader(200)
}

func resumeDownload(id string) {
	defer func() { recover() }()
	task, ok := getDownloadTask(id)
	if !ok {
		return
	}

	sandbox := filepath.Join(config.AppsDir, task.Name)
	fname := filepath.Base(task.URL)
	if fname == "" || fname == "." || fname == "/" {
		fname = "download"
	}
	fpath := filepath.Join(sandbox, fname)

	var startOffset int64
	if info, err := os.Stat(fpath); err == nil {
		startOffset = info.Size()
	}

	client := &http.Client{Timeout: 60 * time.Second}
	req, err := http.NewRequest("GET", task.URL, nil)
	if err != nil {
		task.SetStatus("error")
		task.Last = time.Now()
		return
	}
	if startOffset > 0 {
		req.Header.Set("Range", "bytes="+strconv.FormatInt(startOffset, 10)+"-")
	}
	resp, e := client.Do(req)
	if e != nil {
		task.SetStatus("error")
		task.Last = time.Now()
		return
	}
	defer resp.Body.Close()

	if startOffset > 0 && resp.StatusCode != http.StatusPartialContent {
		_ = os.Remove(fpath)
		startOffset = 0
		task.UpdateProgress(0, 0, 0)
	}

	f, err := os.OpenFile(fpath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		task.SetStatus("error")
		task.Last = time.Now()
		return
	}
	defer f.Close()

	if startOffset == 0 {
		task.UpdateSize(resp.ContentLength)
		task.UpdateProgress(0, 0, 0)
	} else {
		task.UpdateProgress(startOffset, task.GetSize(), 0)
	}

	buf := make([]byte, 32*1024)
	for task.IsDownloading() {
		n, readErr := resp.Body.Read(buf)
		if n > 0 {
			if _, err := f.Write(buf[:n]); err != nil {
				task.SetStatus("error")
				task.Last = time.Now()
				break
			}
			downloaded := task.GetDownloaded() + int64(n)
			size := task.GetSize()
			var prog int
			if size > 0 {
				prog = int(float64(downloaded) / float64(size) * 100)
			}
			task.UpdateProgress(downloaded, size, prog)
		}
		if readErr != nil {
			if !errors.Is(readErr, io.EOF) {
				task.SetStatus("error")
				task.Last = time.Now()
			}
			break
		}
	}

	if task.GetStatus() == "downloading" {
		task.SetStatus("completed")
		task.UpdateProgress(task.GetDownloaded(), task.GetSize(), 100)
	}
}

func installDownloadedApp(id string) {
	task, ok := getDownloadTask(id)
	if !ok {
		return
	}

	appOpMu.Lock()
	defer appOpMu.Unlock()

	sandbox := filepath.Join(config.AppsDir, task.Name)
	fname := filepath.Base(task.URL)
	fpath := filepath.Join(sandbox, fname)

	if extractArchive(fpath, sandbox) {
		_ = os.Remove(fpath)
	} else if strings.HasSuffix(fname, ".sh") {
		_ = os.Chmod(fpath, 0755)
	}

	scanTime := time.Now().Add(-20 * time.Second)
	result := combinedDetect(sandbox, scanTime)

	var apps map[string]models.Project
	_ = LoadJSON(config.ConfigApps, &apps)
	apps[task.Name] = models.Project{
		Path:      sandbox,
		Cmd:       task.Cmd,
		WorkDir:   result.WorkDir,
		Created:   time.Now().Format("2006-01-02 15:04"),
	}
	_ = WriteJSON(config.ConfigApps, apps)
	_ = os.WriteFile(filepath.Join(sandbox, "install_note.txt"), []byte(result.Note), 0644)
}

const maxStoreResponseSize = 10 * 1024 * 1024 // 10MB

func startStoreInstall(w http.ResponseWriter, r *http.Request) {
	idx, _ := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/install/"))
	srcIdx, _ := strconv.Atoi(r.URL.Query().Get("source"))
	var srcs []models.StoreSource
	_ = LoadJSON(config.ConfigSrc, &srcs)
	if srcIdx < 0 || srcIdx >= len(srcs) || idx < 0 {
		http.Redirect(w, r, "/store", 302)
		return
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, e := client.Get(srcs[srcIdx].URL)
	if e != nil {
		http.Redirect(w, r, "/store", 302)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		http.Redirect(w, r, "/store", 302)
		return
	}

	var apps []models.StoreApp
	limitedReader := io.LimitReader(resp.Body, maxStoreResponseSize)
	if err := json.NewDecoder(limitedReader).Decode(&apps); err != nil {
		http.Redirect(w, r, "/store", 302)
		return
	}
	if idx >= len(apps) {
		http.Redirect(w, r, "/store", 302)
		return
	}
	app := apps[idx]
	name := strings.TrimSpace(app.Name)
	if name == "" || strings.ContainsAny(name, `./\ `) || strings.HasPrefix(name, ".") || strings.Contains(name, "..") {
		http.Redirect(w, r, "/store", 302)
		return
	}

	if app.URL == "" {
		appOpMu.Lock()
		var list map[string]models.Project
		_ = LoadJSON(config.ConfigApps, &list)
		if _, exists := list[name]; !exists {
			sandbox := filepath.Join(config.AppsDir, name)
			_ = os.MkdirAll(sandbox, 0755)

			scanTime := time.Now().Add(-20 * time.Second)
			result := combinedDetect(sandbox, scanTime)

			list[name] = models.Project{
				Path:      sandbox,
				Cmd:       app.Cmd,
				WorkDir:   result.WorkDir,
				Version:   app.Version,
				Created:   time.Now().Format("2006-01-02 15:04"),
			}
			_ = WriteJSON(config.ConfigApps, list)
			_ = os.WriteFile(filepath.Join(sandbox, "install_note.txt"), []byte(result.Note), 0644)
		}
		appOpMu.Unlock()
		http.Redirect(w, r, "/", 302)
		return
	}

	if !strings.HasPrefix(app.URL, "http://") && !strings.HasPrefix(app.URL, "https://") {
		http.Redirect(w, r, "/store", 302)
		return
	}
	if isPrivateURL(app.URL) {
		http.Redirect(w, r, "/store", 302)
		return
	}

	dlURL := app.URL
	dlName := name
	dlCmd := app.Cmd
	dlVersion := app.Version
	id := "dl_" + strconv.FormatInt(time.Now().UnixNano(), 36)
	task := &models.DownloadTask{
		ID:     id,
		Name:   dlName,
		URL:    dlURL,
		Cmd:    dlCmd,
		Size:   0,
		Status: "downloading",
		Last:   time.Now(),
	}
	downloadTasks.Store(id, task)

	go func() {
		defer func() { recover() }()
		sandbox := filepath.Join(config.AppsDir, dlName)
		_ = os.MkdirAll(sandbox, 0755)

		fname := filepath.Base(dlURL)
		if fname == "" || fname == "." || fname == "/" {
			fname = "download"
		}
		fpath := filepath.Join(sandbox, fname)

		dlClient := &http.Client{Timeout: 60 * time.Second}
		dlResp, e := dlClient.Get(dlURL)
		if e != nil {
			task.SetStatus("error")
			task.Last = time.Now()
			return
		}
		defer dlResp.Body.Close()

		task.UpdateSize(dlResp.ContentLength)

		f, err := os.Create(fpath)
		if err != nil {
			task.SetStatus("error")
			return
		}

		buf := make([]byte, 32*1024)
		for task.IsDownloading() {
			n, readErr := dlResp.Body.Read(buf)
			if n > 0 {
				if _, err := f.Write(buf[:n]); err != nil {
					task.SetStatus("error")
					task.Last = time.Now()
					break
				}
				downloaded := task.GetDownloaded() + int64(n)
				size := task.GetSize()
				var prog int
				if size > 0 {
					prog = int(float64(downloaded) / float64(size) * 100)
				}
				task.UpdateProgress(downloaded, size, prog)
			}
			if readErr != nil {
				if !errors.Is(readErr, io.EOF) {
					task.SetStatus("error")
					task.Last = time.Now()
				}
				break
			}
		}
		_ = f.Close()

		if task.GetStatus() == "downloading" {
			task.SetStatus("completed")
			task.UpdateProgress(task.GetDownloaded(), task.GetSize(), 100)
			if extractArchive(fpath, sandbox) {
				_ = os.Remove(fpath)
			} else if strings.HasSuffix(fname, ".sh") {
				_ = os.Chmod(fpath, 0755)
			}

			time.Sleep(15 * time.Second)
			scanTime := time.Now().Add(-20 * time.Second)
			result := combinedDetect(sandbox, scanTime)
			appOpMu.Lock()
			var updated map[string]models.Project
			if err := LoadJSON(config.ConfigApps, &updated); err == nil {
				if proj, ok := updated[dlName]; ok {
					if result.WorkDir != "" {
						proj.WorkDir = result.WorkDir
					}
					if result.Binary != "" {
						proj.Cmd = result.Binary
					}
					proj.Version = dlVersion
					updated[dlName] = proj
					_ = WriteJSON(config.ConfigApps, updated)
				}
			}
			appOpMu.Unlock()
			_ = os.WriteFile(filepath.Join(sandbox, "install_note.txt"), []byte(result.Note), 0644)
		}
	}()

	http.Redirect(w, r, "/downloads", 302)
}

func storePage(w http.ResponseWriter, r *http.Request) {
	var srcs []models.StoreSource
	_ = LoadJSON(config.ConfigSrc, &srcs)
	idx, _ := strconv.Atoi(r.URL.Query().Get("source"))
	if idx < 0 || idx >= len(srcs) {
		idx = 0
	}

	var apps []models.StoreApp
	storeErr := ""
	storeErrType := ""
	if len(srcs) > 0 {
		client := &http.Client{Timeout: 10 * time.Second}
		resp, e := client.Get(srcs[idx].URL)
		if e != nil {
			storeErr = "无法连接到商店源"
			storeErrType = "network"
		} else {
			defer resp.Body.Close()
			if resp.StatusCode != 200 {
				storeErr = "商店源返回错误: " + fmt.Sprint(resp.StatusCode)
				storeErrType = "http"
			} else {
				limitedReader := io.LimitReader(resp.Body, maxStoreResponseSize)
				if err := json.NewDecoder(limitedReader).Decode(&apps); err != nil {
					storeErr = "商店数据格式错误"
					storeErrType = "parse"
				}
			}
		}
	}
	if len(apps) == 0 {
		if storeErr == "" {
			storeErr = "没有可用的应用"
		}
		apps = []models.StoreApp{{
			Name:   "获取失败",
			Desc:   storeErr,
			Icon:   "",
			Author: "system",
		}}
	}

	_ = htmlRender.ExecuteTemplate(w, "store", map[string]any{
		"Apps":         apps,
		"Sources":      srcs,
		"Active":       idx,
		"StoreErr":     storeErr,
		"StoreErrType": storeErrType,
	})
}

func checkAppUpdates(apps map[string]models.Project) map[string]string {
	updates := make(map[string]string)

	var srcs []models.StoreSource
	if err := LoadJSON(config.ConfigSrc, &srcs); err != nil || len(srcs) == 0 {
		return updates
	}

	storeAppsByName := make(map[string]map[string]string)
	for _, src := range srcs {
		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Get(src.URL)
		if err != nil {
			continue
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			continue
		}
		var storeApps []models.StoreApp
		if err := json.NewDecoder(io.LimitReader(resp.Body, maxStoreResponseSize)).Decode(&storeApps); err != nil {
			continue
		}
		for _, app := range storeApps {
			if app.Version == "" {
				continue
			}
			if storeAppsByName[app.Name] == nil {
				storeAppsByName[app.Name] = make(map[string]string)
			}
			storeAppsByName[app.Name][src.URL] = app.Version
		}
	}

	for name, app := range apps {
		if app.SourceURL == "" || app.Version == "" {
			continue
		}
		if storeVersions, ok := storeAppsByName[name]; ok {
			if newVer, ok := storeVersions[app.SourceURL]; ok {
				if newVer != app.Version {
					updates[name] = newVer
				}
			}
		}
	}

	return updates
}

	var apps []models.StoreApp
	storeErr := ""
	storeErrType := ""
	if len(srcs) > 0 {
		client := &http.Client{Timeout: 10 * time.Second}
		resp, e := client.Get(srcs[idx].URL)
		if e != nil {
			storeErr = "无法连接到商店源"
			storeErrType = "network"
		} else {
			defer resp.Body.Close()
			if resp.StatusCode != 200 {
				storeErr = "商店源返回错误: " + fmt.Sprint(resp.StatusCode)
				storeErrType = "http"
			} else {
				limitedReader := io.LimitReader(resp.Body, maxStoreResponseSize)
				if err := json.NewDecoder(limitedReader).Decode(&apps); err != nil {
					storeErr = "商店数据格式错误"
					storeErrType = "parse"
				}
			}
		}
	}
	if len(apps) == 0 {
		if storeErr == "" {
			storeErr = "没有可用的应用"
		}
		apps = []models.StoreApp{{
			Name:   "获取失败",
			Desc:   storeErr,
			Icon:   "",
			Author: "system",
		}}
	}

	_ = htmlRender.ExecuteTemplate(w, "store", map[string]any{
		"Apps":        apps,
		"Sources":     srcs,
		"Active":      idx,
		"StoreErr":    storeErr,
		"StoreErrType": storeErrType,
	})
}
