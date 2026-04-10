package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"lightpanel/config"
	"lightpanel/models"
)

const maxStoreResponseSize = 10 * 1024 * 1024

var (
	downloadTasks sync.Map
	storeCache    struct {
		apps      []models.StoreApp
		err       string
		errType   string
		lastInit  time.Time
	}
	storeCacheLock sync.RWMutex
	storeCacheTTL  = 5 * time.Minute
)

func getSystemArch() string {
	arch := runtime.GOARCH
	switch arch {
	case "x86_64":
		return "amd64"
	case "aarch64":
		return "arm64"
	case "armv7l":
		return "armv7"
	case "i386", "i686":
		return "386"
	default:
		return arch
	}
}

func getSystemOS() string {
	os := runtime.GOOS
	switch os {
	case "darwin":
		return "darwin"
	default:
		return "linux"
	}
}

func getCachedSystemInfo() config.SystemInfo {
	var sysInfo config.SystemInfo
	if err := LoadJSON(config.ConfigSys, &sysInfo); err == nil && sysInfo.Arch != "" {
		return sysInfo
	}
	
	arch := getSystemArch()
	os := getSystemOS()
	sysInfo = config.SystemInfo{Arch: arch, OS: os}
	_ = WriteJSON(config.ConfigSys, sysInfo)
	return sysInfo
}

func replaceURLParams(url string) (string, error) {
	sysInfo := getCachedSystemInfo()
	arch := sysInfo.Arch
	os := sysInfo.OS
	
	currentURL := strings.ReplaceAll(url, "{{arch}}", arch)
	currentURL = strings.ReplaceAll(currentURL, "{{os}}", os)
	
	if strings.Contains(currentURL, "{{arch}}") || strings.Contains(currentURL, "{{os}}") {
		return "", fmt.Errorf("当前架构 %s 无可用下载链接", arch)
	}
	
	return currentURL, nil
}

func testURL(url string) bool {
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Head(url)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == 200
}

// isSafeBinary 检查文件是否为安全的可执行文件
// 通过检查文件头部 ELF 魔数和已知安全文件名模式
func isSafeBinary(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()

	header := make([]byte, 4)
	if _, err := f.Read(header); err != nil {
		return false
	}

	// 检查 ELF 魔数: 0x7f 'E' 'L' 'F'
	if header[0] == 0x7f && header[1] == 'E' && header[2] == 'L' && header[3] == 'F' {
		return true
	}

	return false
}

func init() {
	go func() {
		time.Sleep(2 * time.Second)
		var srcs []models.StoreSource
		LoadJSON(config.ConfigSrc, &srcs)
		initStoreCache(srcs, 0)
	}()

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

func initStoreCache(srcs []models.StoreSource, srcIdx int) {
	if len(srcs) == 0 {
		storeCache.apps = nil
		storeCache.err = "没有可用的商店源"
		storeCache.errType = "none"
		storeCache.lastInit = time.Now()
		return
	}

	client := &http.Client{Timeout: 5 * time.Second}
	if srcIdx < 0 || srcIdx >= len(srcs) {
		srcIdx = 0
	}
	resp, err := client.Get(srcs[srcIdx].URL)
	if err != nil {
		storeCache.apps = nil
		storeCache.err = "无法连接到商店源"
		storeCache.errType = "network"
		storeCache.lastInit = time.Now()
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		storeCache.apps = nil
		storeCache.err = fmt.Sprintf("商店源返回错误: %d", resp.StatusCode)
		storeCache.errType = "http"
		storeCache.lastInit = time.Now()
		return
	}

	limitedReader := io.LimitReader(resp.Body, maxStoreResponseSize)
	data, err := io.ReadAll(limitedReader)
	if err != nil {
		storeCache.apps = nil
		storeCache.err = "读取商店数据失败"
		storeCache.errType = "parse"
		storeCache.lastInit = time.Now()
		return
	}

	var apps []models.StoreApp

	if len(data) > 0 && data[0] == '{' {
		var manifest models.StoreManifest
		if err := json.Unmarshal(data, &manifest); err != nil {
			storeCache.apps = nil
			storeCache.err = "商店数据格式错误"
			storeCache.errType = "parse"
			storeCache.lastInit = time.Now()
			return
		}
		apps = manifest.Apps
	} else {
		if err := json.Unmarshal(data, &apps); err != nil {
			storeCache.apps = nil
			storeCache.err = "商店数据格式错误"
			storeCache.errType = "parse"
			storeCache.lastInit = time.Now()
			return
		}
	}

	storeCache.apps = apps
	storeCache.err = ""
	storeCache.errType = ""
	storeCache.lastInit = time.Now()
}

func getStoreApps(srcs []models.StoreSource, srcIdx int) ([]models.StoreApp, string, string) {
	storeCacheLock.RLock()
	if time.Since(storeCache.lastInit) <= storeCacheTTL && len(storeCache.apps) > 0 {
		apps := storeCache.apps
		err := storeCache.err
		errType := storeCache.errType
		storeCacheLock.RUnlock()
		return apps, err, errType
	}
	hasOld := len(storeCache.apps) > 0
	storeCacheLock.RUnlock()

	if hasOld {
		go func() {
			storeCacheLock.Lock()
			defer storeCacheLock.Unlock()
			if time.Since(storeCache.lastInit) > storeCacheTTL {
				initStoreCache(srcs, srcIdx)
			}
		}()
		storeCacheLock.RLock()
		apps := storeCache.apps
		err := storeCache.err
		errType := storeCache.errType
		storeCacheLock.RUnlock()
		return apps, err, errType
	}

	storeCacheLock.Lock()
	defer storeCacheLock.Unlock()

	if time.Since(storeCache.lastInit) <= storeCacheTTL && len(storeCache.apps) > 0 {
		return storeCache.apps, storeCache.err, storeCache.errType
	}

	initStoreCache(srcs, srcIdx)
	return storeCache.apps, storeCache.err, storeCache.errType
}

func startStoreInstall(w http.ResponseWriter, r *http.Request) {
	log.Printf("[install] Received request: %s", r.URL.Path)
	path := strings.TrimPrefix(r.URL.Path, "/install/")
	path = strings.TrimSuffix(path, "/")
	idx, err := strconv.Atoi(path)
	if err != nil {
		log.Printf("[install] Invalid index: %v", err)
		http.Redirect(w, r, "/store", 302)
		return
	}
	srcIdx, _ := strconv.Atoi(r.URL.Query().Get("source"))
	var srcs []models.StoreSource
	_ = LoadJSON(config.ConfigSrc, &srcs)
	if srcIdx < 0 || srcIdx >= len(srcs) || idx < 0 {
		log.Printf("[install] Invalid srcIdx=%d or idx=%d, srcs len=%d", srcIdx, idx, len(srcs))
		http.Redirect(w, r, "/store", 302)
		return
	}

	apps, storeErr, _ := getStoreApps(srcs, srcIdx)
	if storeErr != "" || len(apps) == 0 {
		log.Printf("[install] Store error: %s", storeErr)
		http.Redirect(w, r, "/store?err=store_load", 302)
		return
	}
	if idx >= len(apps) {
		log.Printf("[install] Index %d out of range, apps len=%d", idx, len(apps))
		http.Redirect(w, r, "/store", 302)
		return
	}
	app := apps[idx]
	name := strings.TrimSpace(app.Name)
	log.Printf("[install] Installing app: %s, URL: %s, AskParams: %v", name, app.URL, app.AskParams)
	name = strings.ReplaceAll(name, " ", "-")
	if name == "" || strings.ContainsAny(name, `./\`) || strings.HasPrefix(name, ".") || strings.Contains(name, "..") || strings.ContainsAny(name, ";&|'?*:#") {
		http.Redirect(w, r, "/store?err=invalid_name", 302)
		return
	}

	appOpMu.Lock()
	var list map[string]models.Project
	_ = LoadJSON(config.ConfigApps, &list)
	if _, exists := list[name]; exists {
		appOpMu.Unlock()
		http.Redirect(w, r, "/store?err=already_installed", 302)
		return
	}
	appOpMu.Unlock()

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
		http.Redirect(w, r, "/store?err=invalid_url", 302)
		return
	}
	if isPrivateURL(app.URL) {
		http.Redirect(w, r, "/store?err=private_url", 302)
		return
	}

	if app.AskParams {
		http.Redirect(w, r, "/install/params/"+strconv.Itoa(idx)+"?source="+strconv.Itoa(srcIdx), 302)
		return
	}

	dlURL, urlErr := replaceURLParams(app.URL)
	if urlErr != nil {
		log.Printf("[install] URL error: %v", urlErr)
		http.Redirect(w, r, "/store?err=no_arch", 302)
		return
	}
	dlName := name
	dlCmd := app.Cmd
	dlVersion := app.Version
	dlSetupCmd := app.SetupCmd
	dlAutoExtract := app.AutoExtract
	dlMakeExec := app.MakeExec
	id := "dl_" + strconv.FormatInt(time.Now().UnixNano(), 36)
	log.Printf("[install] Creating task: id=%s, name=%s, url=%s", id, dlName, dlURL)
	task := &models.DownloadTask{
		ID:       id,
		Name:     dlName,
		URL:      dlURL,
		Cmd:      dlCmd,
		Size:     0,
		Status:   "downloading",
		Last:     time.Now(),
		Version:  dlVersion,
		SetupCmd: dlSetupCmd,
		Icon:    app.Icon,
		Port:    app.Port,
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
		task.SetStatus("deploying")
		task.UpdateProgress(task.GetDownloaded(), task.GetSize(), 100)
		extracted := false
		if dlAutoExtract {
			extracted = extractArchive(fpath, sandbox)
			if extracted {
				_ = os.Remove(fpath)
			}
		}
		if !extracted && dlMakeExec {
			if strings.HasSuffix(fname, ".sh") {
				_ = os.Chmod(fpath, 0755)
			} else if isSafeBinary(fpath) {
				_ = os.Chmod(fpath, 0755)
			}
		}

		time.Sleep(15 * time.Second)
		scanTime := time.Now().Add(-20 * time.Second)
		result := combinedDetect(sandbox, scanTime)
		appOpMu.Lock()
		var updated map[string]models.Project
		if err := LoadJSON(config.ConfigApps, &updated); err == nil {
			proj, exists := updated[dlName]
			if !exists {
				proj = models.Project{
					Path:      sandbox,
					Created:   time.Now().Format("2006-01-02 15:04"),
				}
			}
		if result.WorkDir != "" {
			proj.WorkDir = result.WorkDir
		}
		if dlCmd != "" {
			dlCmd = strings.ReplaceAll(dlCmd, "{{arch}}", getCachedSystemInfo().Arch)
			dlCmd = strings.ReplaceAll(dlCmd, "{{os}}", getCachedSystemInfo().OS)
			proj.Cmd = dlCmd
		} else if result.Binary != "" {
			proj.Cmd = result.Binary
		}
		proj.Version = dlVersion
		proj.Icon = task.Icon
		if task.Port > 0 {
			proj.Port = task.Port
			if proj.URL == "" {
				proj.URL = "http://" + getLocalIP() + ":" + strconv.Itoa(task.Port)
			}
		}
		updated[dlName] = proj
		_ = WriteJSON(config.ConfigApps, updated)
	}
	appOpMu.Unlock()
	_ = os.WriteFile(filepath.Join(sandbox, "install_note.txt"), []byte(result.Note), 0644)
	task.SetStatus("completed")
}
	}()

	http.Redirect(w, r, "/downloads", 302)
}

func storePage(w http.ResponseWriter, r *http.Request) {
	cookie, _ := r.Cookie("lp_session")
	var csrfToken string
	if cookie != nil {
		sessData := getSessionData(cookie.Value)
		if sessData != nil {
			csrfToken = sessData.CSRFToken
		}
	}

	var srcs []models.StoreSource
	_ = LoadJSON(config.ConfigSrc, &srcs)
	idx, _ := strconv.Atoi(r.URL.Query().Get("source"))
	if idx < 0 || idx >= len(srcs) {
		idx = 0
	}

	if r.URL.Query().Get("refresh") == "1" {
		storeCacheLock.Lock()
		storeCache.apps = nil
		storeCache.lastInit = time.Time{}
		storeCacheLock.Unlock()
	}

	deployErr := r.URL.Query().Get("err")

	apps, storeErr, storeErrType := getStoreApps(srcs, idx)
	
	if errMsg := r.URL.Query().Get("err"); errMsg == "no_arch" {
		sysInfo := getCachedSystemInfo()
		storeErr = fmt.Sprintf("当前系统架构(%s)无可用下载链接，请检查官方文档", sysInfo.Arch)
		storeErrType = "arch"
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
	var deployed map[string]bool
	if config.ConfigApps != "" {
		var deployedApps map[string]models.Project
		if err := LoadJSON(config.ConfigApps, &deployedApps); err == nil {
			deployed = make(map[string]bool)
			for name := range deployedApps {
				deployed[name] = true
				deployed[strings.ReplaceAll(name, "-", " ")] = true
			}
		}
	}

	_ = htmlRender.ExecuteTemplate(w, "store", map[string]any{
		"Apps":         apps,
		"Sources":      srcs,
		"Active":       idx,
		"StoreErr":     storeErr,
		"StoreErrType": storeErrType,
		"Deployed":     deployed,
		"CSRFToken":    csrfToken,
		"DeployErr":    deployErr,
		"Sidebar":      template.HTML(sidebarHTML("/store")),
		"Topbar":       template.HTML(topbarHTML("应用商店")),
	})
}

func checkAppUpdates(apps map[string]models.Project) map[string]string {
	updates := make(map[string]string)

	var srcs []models.StoreSource
	if err := LoadJSON(config.ConfigSrc, &srcs); err != nil || len(srcs) == 0 {
		return updates
	}

	type storeAppInfo struct {
		name        string
		version     string
		url         string
		autoExtract bool
		makeExec    bool
	}

	appChan := make(chan storeAppInfo, len(srcs)*10)
	var wg sync.WaitGroup

	for _, src := range srcs {
		wg.Add(1)
		go func(srcURL string) {
			defer wg.Done()
			client := &http.Client{Timeout: 5 * time.Second}
			resp, err := client.Get(srcURL)
			if err != nil {
				return
			}
			defer resp.Body.Close()
			if resp.StatusCode != 200 {
				return
			}
			var storeApps []models.StoreApp
			if err := json.NewDecoder(io.LimitReader(resp.Body, maxStoreResponseSize)).Decode(&storeApps); err != nil {
				return
			}
			for _, app := range storeApps {
				if app.Version != "" {
					appChan <- storeAppInfo{name: app.Name, version: app.Version, url: srcURL}
				}
			}
		}(src.URL)
	}

	go func() {
		wg.Wait()
		close(appChan)
	}()

	storeAppsByName := make(map[string]map[string]string)
	for info := range appChan {
		if storeAppsByName[info.name] == nil {
			storeAppsByName[info.name] = make(map[string]string)
		}
		storeAppsByName[info.name][info.url] = info.version
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

func storeParamsPage(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/install/params/")
	path = strings.TrimSuffix(path, "/")
	idx, err := strconv.Atoi(path)
	if err != nil {
		http.Redirect(w, r, "/store", 302)
		return
	}
	srcIdx, _ := strconv.Atoi(r.URL.Query().Get("source"))

	var srcs []models.StoreSource
	_ = LoadJSON(config.ConfigSrc, &srcs)
	if srcIdx < 0 || srcIdx >= len(srcs) || idx < 0 {
		http.Redirect(w, r, "/store", 302)
		return
	}

	apps, _, _ := getStoreApps(srcs, srcIdx)
	if idx >= len(apps) {
		http.Redirect(w, r, "/store", 302)
		return
	}

	app := apps[idx]
	dlURL, _ := replaceURLParams(app.URL)

	_ = htmlRender.ExecuteTemplate(w, "store_params", map[string]any{
		"App":        app,
		"Index":      idx,
		"SrcIdx":     srcIdx,
		"TestedURL":  dlURL,
		"Sidebar":    template.HTML(sidebarHTML("/store")),
		"Topbar":     template.HTML(topbarHTML("设置参数")),
	})
}

func confirmInstallWithParams(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/install/confirm/")
	idx, _ := strconv.Atoi(strings.TrimSuffix(path, "/"))
	srcIdx, _ := strconv.Atoi(r.URL.Query().Get("source"))
	userParams := strings.TrimSpace(r.FormValue("user_params"))

	var srcs []models.StoreSource
	_ = LoadJSON(config.ConfigSrc, &srcs)
	if srcIdx < 0 || srcIdx >= len(srcs) || idx < 0 {
		http.Redirect(w, r, "/store", 302)
		return
	}

	apps, _, _ := getStoreApps(srcs, srcIdx)
	if idx >= len(apps) {
		http.Redirect(w, r, "/store", 302)
		return
	}

	app := apps[idx]
	name := strings.TrimSpace(app.Name)
	if name == "" {
		http.Redirect(w, r, "/store", 302)
		return
	}

	dlURL, urlErr := replaceURLParams(app.URL)
	if urlErr != nil || (dlURL == app.URL && strings.Contains(app.URL, "{{arch}}")) {
		http.Redirect(w, r, "/store?err=no_arch", 302)
		return
	}

	dlName := name
	dlCmd := app.Cmd
	if userParams != "" {
		if !validateCommand(userParams) {
			http.Redirect(w, r, "/store?err=invalid_params", 302)
			return
		}
		if strings.Contains(dlCmd, "{{params}}") {
			dlCmd = strings.ReplaceAll(dlCmd, "{{params}}", userParams)
		} else {
			dlCmd = dlCmd + " " + userParams
		}
	}
	dlCmd = strings.ReplaceAll(dlCmd, "{{arch}}", getCachedSystemInfo().Arch)
	dlCmd = strings.ReplaceAll(dlCmd, "{{os}}", getCachedSystemInfo().OS)
	dlVersion := app.Version
	dlSetupCmd := app.SetupCmd
	dlWorkDir := app.WorkDir
	id := "dl_" + strconv.FormatInt(time.Now().UnixNano(), 36)
	task := &models.DownloadTask{
		ID:       id,
		Name:     dlName,
		URL:      dlURL,
		Cmd:      dlCmd,
		Size:     0,
		Status:   "downloading",
		Last:     time.Now(),
		Version:  dlVersion,
		SetupCmd: dlSetupCmd,
		WorkDir:  dlWorkDir,
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
			task.SetStatus("deploying")
			task.UpdateProgress(task.GetDownloaded(), task.GetSize(), 100)
			extracted := false
			if app.AutoExtract {
				extracted = extractArchive(fpath, sandbox)
				if extracted {
					_ = os.Remove(fpath)
				}
			}
			if !extracted && app.MakeExec {
				if strings.HasSuffix(fname, ".sh") {
					_ = os.Chmod(fpath, 0755)
				} else if isSafeBinary(fpath) {
					_ = os.Chmod(fpath, 0755)
				}
		}

		time.Sleep(15 * time.Second)
		scanTime := time.Now().Add(-20 * time.Second)
		result := combinedDetect(sandbox, scanTime)
		appOpMu.Lock()
		var updated map[string]models.Project
		if err := LoadJSON(config.ConfigApps, &updated); err == nil {
			proj, exists := updated[dlName]
			if !exists {
				proj = models.Project{
					Path:      sandbox,
					Created:   time.Now().Format("2006-01-02 15:04"),
				}
			}
		if result.WorkDir != "" {
			proj.WorkDir = result.WorkDir
		}
		if dlCmd != "" {
			dlCmd = strings.ReplaceAll(dlCmd, "{{arch}}", getCachedSystemInfo().Arch)
			dlCmd = strings.ReplaceAll(dlCmd, "{{os}}", getCachedSystemInfo().OS)
			proj.Cmd = dlCmd
} else if result.Binary != "" {
proj.Cmd = result.Binary
		}
		proj.Version = dlVersion
		proj.Icon = task.Icon
		if task.Port > 0 {
			proj.Port = task.Port
			if proj.URL == "" {
				proj.URL = "http://" + getLocalIP() + ":" + strconv.Itoa(task.Port)
			}
		}
		updated[dlName] = proj
			_ = WriteJSON(config.ConfigApps, updated)
	}
	appOpMu.Unlock()
	_ = os.WriteFile(filepath.Join(sandbox, "install_note.txt"), []byte(result.Note), 0644)
	task.SetStatus("completed")
}
}()

	http.Redirect(w, r, "/downloads", 302)
}

func apiCheckUpdates(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var apps map[string]models.Project
	_ = LoadJSON(config.ConfigApps, &apps)

	updates := checkAppUpdates(apps)
	json.NewEncoder(w).Encode(updates)
}

func downloadPage(w http.ResponseWriter, r *http.Request) {
	cookie, _ := r.Cookie("lp_session")
	var csrfToken string
	if cookie != nil {
		sessData := getSessionData(cookie.Value)
		if sessData != nil {
			csrfToken = sessData.CSRFToken
		}
	}

	var tasks []*models.DownloadTask
	downloadTasks.Range(func(key, value any) bool {
		if t, ok := value.(*models.DownloadTask); ok {
			tasks = append(tasks, t)
		}
		return true
	})

	_ = htmlRender.ExecuteTemplate(w, "downloads", map[string]any{
		"Tasks":      tasks,
		"History":    nil,
		"Sidebar":    template.HTML(sidebarHTML("/downloads")),
		"Topbar":     template.HTML(topbarHTML("下载管理")),
		"CSRFToken":  csrfToken,
	})
}

func apiDownloads(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var tasks []map[string]any
	downloadTasks.Range(func(key, value any) bool {
		if t, ok := value.(*models.DownloadTask); ok {
			tasks = append(tasks, map[string]any{
				"id":         t.ID,
				"name":       t.Name,
				"url":        t.URL,
				"size":       t.GetSize(),
				"downloaded": t.GetDownloaded(),
				"status":     t.GetStatus(),
				"progress":   t.GetProgress(),
			})
		}
		return true
	})
	json.NewEncoder(w).Encode(tasks)
}

func apiDownloadAction(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/dl/action/")
	parts := strings.SplitN(path, "/", 2)
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

func getDownloadTask(id string) (*models.DownloadTask, bool) {
	v, ok := downloadTasks.Load(id)
	if !ok {
		return nil, false
	}
	t, ok := v.(*models.DownloadTask)
	return t, ok
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

	if startOffset > 0 && resp.StatusCode == http.StatusOK {
		expectedSize := task.GetSize()
		actualSize := resp.ContentLength
		if actualSize > 0 && expectedSize > 0 && actualSize != expectedSize {
			_ = os.Remove(fpath)
			startOffset = 0
			task.UpdateProgress(0, 0, 0)
		}
	}

	var openMode int
	if startOffset > 0 {
		openMode = os.O_CREATE|os.O_WRONLY|os.O_APPEND
	} else {
		openMode = os.O_CREATE|os.O_WRONLY|os.O_TRUNC
	}
	f, err := os.OpenFile(fpath, openMode, 0644)
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
			if task.GetStatus() == "paused" {
				break
			}
			if !errors.Is(readErr, io.EOF) {
				task.SetStatus("error")
				task.Last = time.Now()
			}
			break
		}
	}

	if task.GetStatus() == "downloading" {
		task.SetStatus("deploying")
		task.UpdateProgress(task.GetDownloaded(), task.GetSize(), 100)
		extracted := false
		if task.AutoExtract {
			extracted = extractArchive(fpath, sandbox)
			if extracted {
				_ = os.Remove(fpath)
			}
		}
		if !extracted && task.MakeExec {
			if strings.HasSuffix(fname, ".sh") {
				_ = os.Chmod(fpath, 0755)
			} else if isSafeBinary(fpath) {
				_ = os.Chmod(fpath, 0755)
			}
		}

		time.Sleep(15 * time.Second)
		scanTime := time.Now().Add(-20 * time.Second)
		result := combinedDetect(sandbox, scanTime)
		appOpMu.Lock()
		var updated map[string]models.Project
dlName := task.Name
		dlCmd := task.Cmd
		dlVersion := task.Version
		dlIcon := task.Icon
		dlPort := task.Port
		if err := LoadJSON(config.ConfigApps, &updated); err == nil {
			proj, exists := updated[task.Name]
			if !exists {
				proj = models.Project{
					Path:      sandbox,
					Created:   time.Now().Format("2006-01-02 15:04"),
				}
			}
			if result.WorkDir != "" {
				proj.WorkDir = result.WorkDir
			}
			if dlCmd != "" {
				dlCmd = strings.ReplaceAll(dlCmd, "{{arch}}", getCachedSystemInfo().Arch)
				dlCmd = strings.ReplaceAll(dlCmd, "{{os}}", getCachedSystemInfo().OS)
				proj.Cmd = dlCmd
			} else if result.Binary != "" {
proj.Cmd = result.Binary
		}
		proj.Version = dlVersion
		proj.Icon = dlIcon
		if dlPort > 0 {
			proj.Port = dlPort
			if proj.URL == "" {
				proj.URL = "http://" + getLocalIP() + ":" + strconv.Itoa(dlPort)
			}
		}
		updated[dlName] = proj
			_ = WriteJSON(config.ConfigApps, updated)
		}
	appOpMu.Unlock()
	_ = os.WriteFile(filepath.Join(sandbox, "install_note.txt"), []byte(result.Note), 0644)
		task.SetStatus("completed")
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
		SetupCmd:  task.SetupCmd,
		WorkDir:   result.WorkDir,
		Created:   time.Now().Format("2006-01-02 15:04"),
	}
	_ = WriteJSON(config.ConfigApps, apps)
	_ = os.WriteFile(filepath.Join(sandbox, "install_note.txt"), []byte(result.Note), 0644)
}
