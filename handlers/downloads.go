package handlers

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"lightpanel/config"
	"lightpanel/models"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

const maxStoreResponseSize = 10 * 1024 * 1024

var downloadTasks sync.Map
var storeCache struct {
	apps              []models.StoreApp
	err, errType, key string
	lastInit          time.Time
}
var storeCacheLock sync.RWMutex
var storeCacheTTL = 5 * time.Minute

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
	client := remoteClient(3 * time.Second)
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

func initStoreCache(srcs []models.StoreSource, idx int) {
	// Caller holds storeCacheLock; network requests have bounded time and size.
	storeCache.apps = nil
	storeCache.err = ""
	storeCache.errType = ""
	storeCache.key = ""
	storeCache.lastInit = time.Now()
	if idx < 0 || idx >= len(srcs) {
		storeCache.err = "没有可用的商店源"
		storeCache.errType = "none"
		return
	}
	storeCache.key = srcs[idx].URL
	apps, err := fetchStoreApps(srcs[idx].URL)
	if err != nil {
		storeCache.err = err.Error()
		storeCache.errType = "network"
		return
	}
	storeCache.apps = apps
}
func fetchStoreApps(raw string) ([]models.StoreApp, error) {
	resp, err := remoteClient(10 * time.Second).Get(raw)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("商店源返回 HTTP %d", resp.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, maxStoreResponseSize+1))
	if err != nil {
		return nil, err
	}
	if len(data) > maxStoreResponseSize {
		return nil, fmt.Errorf("商店数据过大")
	}
	data = []byte(strings.TrimSpace(string(data)))
	if len(data) == 0 {
		return nil, fmt.Errorf("商店数据为空")
	}
	var apps []models.StoreApp
	if data[0] == '{' {
		var manifest models.StoreManifest
		err = json.Unmarshal(data, &manifest)
		apps = manifest.Apps
	} else {
		err = json.Unmarshal(data, &apps)
	}
	if err != nil {
		return nil, fmt.Errorf("商店数据格式错误")
	}
	return apps, nil
}
func getStoreApps(srcs []models.StoreSource, idx int) ([]models.StoreApp, string, string) {
	storeCacheLock.Lock()
	defer storeCacheLock.Unlock()
	key := ""
	if idx >= 0 && idx < len(srcs) {
		key = srcs[idx].URL
	}
	if storeCache.key != key || time.Since(storeCache.lastInit) > storeCacheTTL {
		initStoreCache(srcs, idx)
	}
	return append([]models.StoreApp(nil), storeCache.apps...), storeCache.err, storeCache.errType
}
func storeDigest(source string, app models.StoreApp) string {
	b, _ := json.Marshal(app)
	sum := sha256.Sum256(append([]byte(source+"\n"), b...))
	return hex.EncodeToString(sum[:])
}
func selectedStoreApp(r *http.Request, prefix string) (models.StoreApp, string, int, int, error) {
	idx, err := strconv.Atoi(strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, prefix), "/"))
	if err != nil {
		return models.StoreApp{}, "", 0, 0, fmt.Errorf("应用索引无效")
	}
	srcIdx, err := strconv.Atoi(r.URL.Query().Get("source"))
	if err != nil {
		return models.StoreApp{}, "", 0, 0, fmt.Errorf("来源无效")
	}
	var srcs []models.StoreSource
	if LoadJSON(config.ConfigSrc, &srcs) != nil || srcIdx < 0 || srcIdx >= len(srcs) {
		return models.StoreApp{}, "", 0, 0, fmt.Errorf("来源不存在")
	}
	apps, msg, _ := getStoreApps(srcs, srcIdx)
	if msg != "" || idx < 0 || idx >= len(apps) {
		return models.StoreApp{}, "", 0, 0, fmt.Errorf("应用列表已变化，请刷新商店")
	}
	return apps[idx], srcs[srcIdx].URL, idx, srcIdx, nil
}
func storeError(w http.ResponseWriter, r *http.Request, err error) {
	http.Redirect(w, r, "/store?err="+url.QueryEscape(err.Error()), http.StatusSeeOther)
}
func storePlan(app models.StoreApp, source, params string) (installPlan, error) {
	var err error
	if app.URL != "" {
		app.URL, err = replaceURLParams(app.URL)
		if err != nil {
			return installPlan{}, err
		}
	}
	sys := getCachedSystemInfo()
	expand := func(s string) string { return strings.NewReplacer("{{arch}}", sys.Arch, "{{os}}", sys.OS).Replace(s) }
	app.Cmd = expand(app.Cmd)
	app.SetupCmd = expand(app.SetupCmd)
	app.WorkDir = expand(app.WorkDir)
	if params != "" {
		app.Cmd = strings.TrimSpace(app.Cmd + " " + params)
	}
	return installPlan{App: app, Source: source}, nil
}
func startStoreInstall(w http.ResponseWriter, r *http.Request) {
	app, src, idx, srcIdx, err := selectedStoreApp(r, "/install/")
	if err == nil && r.FormValue("manifest") != storeDigest(src, app) {
		err = fmt.Errorf("应用列表已变化，请刷新后确认")
	}
	if err != nil {
		storeError(w, r, err)
		return
	}
	if app.AskParams {
		http.Redirect(w, r, fmt.Sprintf("/install/params/%d?source=%d&manifest=%s", idx, srcIdx, storeDigest(src, app)), 303)
		return
	}
	plan, err := storePlan(app, src, "")
	if err == nil {
		_, err = submitInstall(plan, nil)
	}
	if err != nil {
		storeError(w, r, err)
		return
	}
	http.Redirect(w, r, "/downloads", 303)
}
func storeParamsPage(w http.ResponseWriter, r *http.Request) {
	app, src, idx, srcIdx, err := selectedStoreApp(r, "/install/params/")
	digest := storeDigest(src, app)
	if previous := r.URL.Query().Get("manifest"); previous != "" && previous != digest {
		err = fmt.Errorf("应用列表已变化，请重新选择")
	}
	if err != nil {
		storeError(w, r, err)
		return
	}
	plan, err := storePlan(app, src, "")
	if err != nil {
		storeError(w, r, err)
		return
	}
	_ = htmlRender.ExecuteTemplate(w, "store_params", map[string]any{"App": app, "Index": idx, "SrcIdx": srcIdx, "Manifest": digest, "TestedURL": plan.App.URL, "Sidebar": template.HTML(sidebarHTML("/store")), "Topbar": template.HTML(topbarHTML("设置参数"))})
}
func confirmInstallWithParams(w http.ResponseWriter, r *http.Request) {
	app, src, _, _, err := selectedStoreApp(r, "/install/confirm/")
	if err == nil && r.FormValue("manifest") != storeDigest(src, app) {
		err = fmt.Errorf("应用列表已变化，请刷新后确认")
	}
	if err != nil {
		storeError(w, r, err)
		return
	}
	plan, err := storePlan(app, src, strings.TrimSpace(r.FormValue("user_params")))
	if err == nil {
		_, err = submitInstall(plan, nil)
	}
	if err != nil {
		storeError(w, r, err)
		return
	}
	http.Redirect(w, r, "/downloads", 303)
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

	manifests := make([]string, len(apps))
	if idx >= 0 && idx < len(srcs) {
		for i, a := range apps {
			manifests[i] = storeDigest(srcs[idx].URL, a)
		}
	}
	_ = htmlRender.ExecuteTemplate(w, "store", map[string]any{
		"Apps":         apps,
		"Manifests":    manifests,
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
	updates := map[string]string{}
	var srcs []models.StoreSource
	if LoadJSON(config.ConfigSrc, &srcs) != nil {
		return updates
	}
	for idx, src := range srcs {
		entries, msg, _ := getStoreApps(srcs, idx)
		if msg != "" {
			continue
		}
		for _, a := range entries {
			n, err := normalizedAppName(a.Name)
			if err != nil {
				continue
			}
			for name, p := range apps {
				pn, _ := normalizedAppName(name)
				if pn == n && p.SourceURL == src.URL && p.Version != "" && a.Version != "" && a.Version != p.Version {
					updates[name] = a.Version
				}
			}
		}
	}
	return updates
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
			snapshot := t.Snapshot()
			tasks = append(tasks, &snapshot)
		}
		return true
	})

	_ = htmlRender.ExecuteTemplate(w, "downloads", map[string]any{
		"Tasks":     tasks,
		"History":   nil,
		"Sidebar":   template.HTML(sidebarHTML("/downloads")),
		"Topbar":    template.HTML(topbarHTML("下载管理")),
		"CSRFToken": csrfToken,
	})
}

func apiDownloads(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	tasks := []models.DownloadTask{}
	downloadTasks.Range(func(key, value any) bool {
		if t, ok := value.(*models.DownloadTask); ok {
			tasks = append(tasks, t.Snapshot())
		}
		return true
	})
	json.NewEncoder(w).Encode(tasks)
}
func getDownloadTask(id string) (*models.DownloadTask, bool) {
	v, ok := downloadTasks.Load(id)
	if !ok {
		return nil, false
	}
	t, ok := v.(*models.DownloadTask)
	return t, ok
}
func apiDownloadAction(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/dl/action/"), "/")
	if len(parts) != 2 {
		http.Error(w, "无效操作", 400)
		return
	}
	v, ok := installJobs.Load(parts[0])
	if !ok {
		http.Error(w, "任务不存在", 404)
		return
	}
	if err := v.(*installJob).action(parts[1]); err != nil {
		http.Error(w, err.Error(), 409)
		return
	}
	w.WriteHeader(204)
}
