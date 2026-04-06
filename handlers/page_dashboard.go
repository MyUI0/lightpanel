package handlers

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/process"

	"lightpanel/config"
	"lightpanel/models"
)

var (
	cpuFirst     = true
	cpuFirstLock sync.Mutex
	cpuPercent   int32
)

func init() {
	go func() {
		val, _ := cpu.Percent(time.Second, false)
		cpuFirstLock.Lock()
		if len(val) > 0 {
			cpuPercent = int32(val[0])
		}
		cpuFirst = false
		cpuFirstLock.Unlock()
	}()
}

func getCpuPercent() int {
	cpuFirstLock.Lock()
	if cpuFirst {
		cpuFirstLock.Unlock()
		val, _ := cpu.Percent(0, false)
		if len(val) > 0 {
			cpuPercent = int32(val[0])
		}
		return int(cpuPercent)
	}
	cpuFirstLock.Unlock()
	
	val, _ := cpu.Percent(0, false)
	if len(val) > 0 {
		cpuPercent = int32(val[0])
	}
	return int(cpuPercent)
}

func autoStartApps() {
	appOpMu.Lock()
	apps := make(map[string]models.Project)
	if err := LoadJSON(config.ConfigApps, &apps); err != nil {
		appOpMu.Unlock()
		return
	}
	appOpMu.Unlock()

	time.Sleep(1 * time.Second)

	for name, app := range apps {
		if app.AutoStart {
			appOpMu.Lock()
			if launchApp(name, &app) {
				log.Printf("[auto-start] 已启动: %s", name)
			}
			appOpMu.Unlock()
		}
	}
}

type FailInfo struct {
	Name  string
	Log   string
	Deps  []string
	Count int
}

var depPatterns = map[string][]string{
	"libssl":    {"libssl.so", "libcrypto.so"},
	"libcurl":   {"libcurl.so"},
	"libstdc++": {"libstdc++.so", "GLIBCXX"},
	"glibc":     {"GLIBC_", "libc.so"},
	"node":      {"node: not found", "command not found: node"},
	"python":    {"python: not found", "python3: not found"},
	"java":      {"java: not found", "JAVA_HOME"},
	"npm":       {"npm: not found"},
	"docker":    {"docker: not found", "Cannot connect to Docker"},
	"permission": {"permission denied", "EACCES"},
}

func detectDeps(logContent string) []string {
	var deps []string
	for dep, patterns := range depPatterns {
		for _, p := range patterns {
			if strings.Contains(logContent, p) {
				deps = append(deps, dep)
				break
			}
		}
	}
	return deps
}

var failInfoCache struct {
	info    *FailInfo
	ts     time.Time
	mu     sync.Mutex
}

func getFailInfo(apps map[string]models.Project) *FailInfo {
	failInfoCache.mu.Lock()
	defer failInfoCache.mu.Unlock()

	if time.Since(failInfoCache.ts) < 30*time.Second && failInfoCache.info != nil {
		return failInfoCache.info
	}

	var firstFail *FailInfo
	count := 0
	for name, app := range apps {
		pidFile := filepath.Join(app.Path, "pid.pid")
		if _, e := os.Stat(pidFile); os.IsNotExist(e) {
			logFile := filepath.Join(app.Path, "run.log")
			if b, err := os.ReadFile(logFile); err == nil {
				content := string(b)
				lines := strings.Split(content, "\n")
				if len(lines) > 50 {
					lines = lines[len(lines)-50:]
				}
				recentLog := strings.Join(lines, "\n")

				hasError := false
				for _, kw := range []string{"error", "fail", "fatal", "panic", "exception", "crash", "cannot", "not found", "no such", "refused", "denied", "missing"} {
					if strings.Contains(strings.ToLower(recentLog), kw) {
						hasError = true
						break
					}
				}
				if hasError {
					count++
					if firstFail == nil {
						firstFail = &FailInfo{
							Name:  name,
							Log:   recentLog,
							Deps:  detectDeps(recentLog),
						}
					}
				}
			}
		}
	}
	if firstFail != nil {
		firstFail.Count = count
	}

	failInfoCache.info = firstFail
	failInfoCache.ts = time.Now()

	return firstFail
}

func indexPage(w http.ResponseWriter, r *http.Request) {
	apps := make(map[string]models.Project)
	_ = LoadJSON(config.ConfigApps, &apps)

	list := map[string]models.Project{}
	for name, app := range apps {
		app.Name = name
		checkAppStatus(&app)
		list[name] = app
	}

	cpuVal := getCpuPercent()
	memInfo, _ := mem.VirtualMemory()
	diskInfo, _ := disk.Usage("/")
	procs, _ := process.Processes()
	uptime, _ := host.Uptime()

	var uptimeStr string
	if uptime > 86400 {
		uptimeStr = fmt.Sprintf("%d天%d时", uptime/86400, (uptime%86400)/3600)
	} else if uptime > 3600 {
		uptimeStr = fmt.Sprintf("%d时%d分", uptime/3600, (uptime%3600)/60)
	} else if uptime > 60 {
		uptimeStr = fmt.Sprintf("%d分", uptime/60)
	} else {
		uptimeStr = fmt.Sprintf("%d秒", uptime)
	}

	var bgConfig models.BgConfig
	_ = LoadJSON(config.ConfigBg, &bgConfig)

	memVal := 0
	if memInfo != nil {
		memVal = int(memInfo.UsedPercent)
	}
	diskVal := 0
	if diskInfo != nil {
		diskVal = int(diskInfo.UsedPercent)
	}

	createErr := ""
	if r.URL.Query().Get("err") == "download" {
		createErr = r.URL.Query().Get("msg")
	}

	_ = htmlRender.ExecuteTemplate(w, "index", map[string]any{
		"Apps":       list,
		"Cpu":        cpuVal,
		"Mem":        memVal,
		"Disk":       diskVal,
		"ProcNum":    len(procs),
		"Uptime":     uptimeStr,
		"BgUrl":      bgConfig.URL,
		"Sidebar":    template.HTML(sidebarHTML("/")),
		"Topbar":     template.HTML(topbarHTML("应用管理")),
		"FailInfo":   getFailInfo(list),
		"CreateErr":  createErr,
		"FirstLogin": isFirstLogin(r),
	})
}

func systemPage(w http.ResponseWriter, r *http.Request) {
	var list []models.ProcInfo
	ps, _ := process.Processes()
	pidSet := make(map[int32]bool)
	for _, p := range ps {
		if pidSet[p.Pid] {
			continue
		}
		pidSet[p.Pid] = true
		n, _ := p.Name()
		cpuVal, _ := p.CPUPercent()
		memVal, _ := p.MemoryPercent()
		list = append(list, models.ProcInfo{PID: p.Pid, Name: n, Cpu: cpuVal, Mem: float64(memVal)})
	}

	cpuVal := getCpuPercent()
	memInfo, _ := mem.VirtualMemory()
	diskInfo, _ := disk.Usage("/")
	memVal := 0
	if memInfo != nil {
		memVal = int(memInfo.UsedPercent)
	}
	diskVal := 0
	if diskInfo != nil {
		diskVal = int(diskInfo.UsedPercent)
	}

	_ = htmlRender.ExecuteTemplate(w, "system", map[string]any{
		"Cpu":     cpuVal,
		"Mem":     memVal,
		"Disk":    diskVal,
		"ProcNum": len(list),
		"Procs":   list,
		"Sidebar": template.HTML(sidebarHTML("/system")),
		"Topbar":  template.HTML(topbarHTML("系统监控")),
	})
}
