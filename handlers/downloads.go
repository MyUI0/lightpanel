package handlers

import (
	"encoding/json"
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
	go func() {
		for range time.Tick(10 * time.Minute) {
			downloadTasks.Range(func(key, value any) bool {
				if t, ok := value.(*models.DownloadTask); ok {
					if t.Status == "completed" || t.Status == "error" {
						if time.Since(t.last) > 30*time.Minute {
							downloadTasks.Delete(key)
						}
					}
				}
				return true
			})
		}
	}()
}

func downloadPage(w http.ResponseWriter, r *http.Request) {
	var tasks []models.DownloadTask
	downloadTasks.Range(func(_, value any) bool {
		if t, ok := value.(*models.DownloadTask); ok {
			tasks = append(tasks, *t)
		}
		return true
	})
	_ = htmlRender.ExecuteTemplate(w, "downloads", tasks)
}

func apiDownloads(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var tasks []models.DownloadTask
	downloadTasks.Range(func(_, value any) bool {
		if t, ok := value.(*models.DownloadTask); ok {
			tasks = append(tasks, *t)
		}
		return true
	})
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

	val, ok := downloadTasks.Load(id)
	if !ok {
		w.WriteHeader(404)
		return
	}
	task := val.(*models.DownloadTask)

	switch act {
	case "pause":
		if task.Status == "downloading" {
			task.Status = "paused"
		}
	case "resume":
		if task.Status == "paused" {
			task.Status = "downloading"
			go resumeDownload(id)
		}
	case "delete":
		downloadTasks.Delete(id)
		if task.Status == "downloading" || task.Status == "paused" {
			fpath := filepath.Join(config.AppsDir, task.Name, filepath.Base(task.URL))
			os.Remove(fpath)
		}
	case "install":
		if task.Status == "completed" {
			installDownloadedApp(id)
			downloadTasks.Delete(id)
		}
	}

	w.WriteHeader(200)
}

func resumeDownload(id string) {
	val, ok := downloadTasks.Load(id)
	if !ok {
		return
	}
	task := val.(*models.DownloadTask)

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
	req, _ := http.NewRequest("GET", task.URL, nil)
	if startOffset > 0 {
		req.Header.Set("Range", "bytes="+strconv.FormatInt(startOffset, 10)+"-")
	}
	resp, e := client.Do(req)
	if e != nil {
		task.Status = "error"
		task.last = time.Now()
		return
	}
	defer resp.Body.Close()

	f, err := os.OpenFile(fpath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		task.Status = "error"
		task.last = time.Now()
		return
	}
	defer f.Close()

	if startOffset == 0 {
		task.Size = resp.ContentLength
		if task.Size <= 0 {
			task.Size = 0
		}
	} else {
		task.Downloaded = startOffset
	}

	for task.Status == "downloading" {
		buf := make([]byte, 32*1024)
		n, readErr := resp.Body.Read(buf)
		if n > 0 {
			f.Write(buf[:n])
			task.Downloaded += int64(n)
			if task.Size > 0 {
				task.Progress = int(float64(task.Downloaded) / float64(task.Size) * 100)
			}
		}
		if readErr != nil {
			break
		}
	}

	if task.Status == "downloading" {
		task.Status = "completed"
		task.Progress = 100
		task.last = time.Now()
	}
}

func installDownloadedApp(id string) {
	val, ok := downloadTasks.Load(id)
	if !ok {
		return
	}
	task := val.(*models.DownloadTask)

	sandbox := filepath.Join(config.AppsDir, task.Name)
	fname := filepath.Base(task.URL)
	fpath := filepath.Join(sandbox, fname)

	if strings.HasSuffix(fname, ".sh") {
		os.Chmod(fpath, 0755)
	}

	var apps map[string]models.Project
	_ = LoadJSON(config.ConfigApps, &apps)
	apps[task.Name] = models.Project{
		Path:    sandbox,
		Cmd:     task.Cmd,
		Created: time.Now().Format("2006-01-02 15:04"),
	}
	_ = WriteJSON(config.ConfigApps, apps)
}

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
	if e != nil || resp.StatusCode != 200 {
		http.Redirect(w, r, "/store", 302)
		return
	}
	defer resp.Body.Close()

	var apps []models.StoreApp
	_ = json.NewDecoder(resp.Body).Decode(&apps)
	if idx >= len(apps) {
		http.Redirect(w, r, "/store", 302)
		return
	}
	app := apps[idx]
	name := strings.TrimSpace(app.Name)
	if strings.ContainsAny(name, `./\ `) {
		http.Redirect(w, r, "/store", 302)
		return
	}

	if app.URL == "" {
		sandbox := filepath.Join(config.AppsDir, name)
		os.MkdirAll(sandbox, 0755)
		var list map[string]models.Project
		_ = LoadJSON(config.ConfigApps, &list)
		list[name] = models.Project{
			Path:    sandbox,
			Cmd:     app.Cmd,
			Created: time.Now().Format("2006-01-02 15:04"),
		}
		_ = WriteJSON(config.ConfigApps, list)
		http.Redirect(w, r, "/", 302)
		return
	}

	id := "dl_" + strconv.FormatInt(time.Now().UnixNano(), 36)
	task := &models.DownloadTask{
		ID:   id,
		Name: name,
		URL:  app.URL,
		Cmd:  app.Cmd,
		Size: 0,
		Status: "downloading",
		last: time.Now(),
	}
	downloadTasks.Store(id, task)

	go func() {
		sandbox := filepath.Join(config.AppsDir, name)
		os.MkdirAll(sandbox, 0755)

		fname := filepath.Base(app.URL)
		if fname == "" || fname == "." || fname == "/" {
			fname = "download"
		}
		fpath := filepath.Join(sandbox, fname)

		dlClient := &http.Client{Timeout: 60 * time.Second}
		dlResp, e := dlClient.Get(app.URL)
		if e != nil {
			task.Status = "error"
			task.last = time.Now()
			return
		}
		defer dlResp.Body.Close()

		task.Size = dlResp.ContentLength

		f, err := os.Create(fpath)
		if err != nil {
			task.Status = "error"
			return
		}

		buf := make([]byte, 32*1024)
		for task.Status == "downloading" {
			n, readErr := dlResp.Body.Read(buf)
			if n > 0 {
				f.Write(buf[:n])
				task.Downloaded += int64(n)
				if task.Size > 0 {
					task.Progress = int(float64(task.Downloaded) / float64(task.Size) * 100)
				}
			}
			if readErr != nil {
				break
			}
		}
		f.Close()

		if task.Status == "downloading" {
			task.Status = "completed"
			task.Progress = 100
			task.last = time.Now()
			if strings.HasSuffix(fname, ".sh") {
				os.Chmod(fpath, 0755)
			}
			var list map[string]models.Project
			_ = LoadJSON(config.ConfigApps, &list)
			list[name] = models.Project{
				Path:    sandbox,
				Cmd:     app.Cmd,
				Created: time.Now().Format("2006-01-02 15:04"),
			}
			_ = WriteJSON(config.ConfigApps, list)
		}
	}()

	http.Redirect(w, r, "/downloads", 302)
}
