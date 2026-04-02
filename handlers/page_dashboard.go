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

var htmlRender *template.Template

var (
	cpuFirst     = true
	cpuFirstLock sync.Mutex
)

func getCpuPercent() int {
	cpuFirstLock.Lock()
	if cpuFirst {
		cpuFirst = false
		cpuFirstLock.Unlock()
		val, _ := cpu.Percent(time.Second, false)
		if len(val) > 0 {
			return int(val[0])
		}
		return 0
	}
	cpuFirstLock.Unlock()

	val, _ := cpu.Percent(0, false)
	if len(val) > 0 {
		return int(val[0])
	}
	return 0
}

func InitTemplates() {
	htmlRender = template.Must(template.New("web").Funcs(template.FuncMap{
		"formatFloat": func(f float64) string { return fmt.Sprintf("%.1f", f) },
		"formatSize": func(b int64) string {
			if b <= 0 {
				return "0 B"
			}
			u := []string{"B", "KB", "MB", "GB"}
			i := 0
			v := float64(b)
			for v >= 1024 && i < len(u)-1 {
				v /= 1024
				i++
			}
			return fmt.Sprintf("%.1f %s", v, u[i])
		},
	}).Parse(`{{define "index"}}`+htmlIndex+`{{end}}`+
		`{{define "store"}}`+htmlStore+`{{end}}`+
		`{{define "source"}}`+htmlSource+`{{end}}`+
		`{{define "system"}}`+htmlSystem+`{{end}}`+
		`{{define "log"}}`+htmlLog+`{{end}}`+
		`{{define "setting"}}`+htmlSetting+`{{end}}`+
		`{{define "edit"}}`+htmlEdit+`{{end}}`+
		`{{define "downloads"}}`+htmlDownloads+`{{end}}`))
}

func autoStartApps() {
	apps := make(map[string]models.Project)
	if err := LoadJSON(config.ConfigApps, &apps); err != nil {
		return
	}

	time.Sleep(2 * time.Second)

	for name, app := range apps {
		if app.AutoStart {
			if launchApp(name, app) {
				log.Printf("[auto-start] 已启动: %s", name)
			}
		}
	}
}

type FailInfo struct {
	Name string
	Log  string
	Deps []string
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

func getFailInfo(apps map[string]models.Project) *FailInfo {
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
					return &FailInfo{
						Name: name,
						Log:  recentLog,
						Deps: detectDeps(recentLog),
					}
				}
			}
		}
	}
	return nil
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
	} else {
		uptimeStr = fmt.Sprintf("%d分", uptime/60)
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

	_ = htmlRender.ExecuteTemplate(w, "index", map[string]any{
		"Apps":    list,
		"Cpu":     cpuVal,
		"Mem":     memVal,
		"Disk":    diskVal,
		"ProcNum": len(procs),
		"Uptime":  uptimeStr,
		"BgUrl":   bgConfig.URL,
		"FailInfo": getFailInfo(apps),
	})
}

func systemPage(w http.ResponseWriter, r *http.Request) {
	var list []models.ProcInfo
	ps, _ := process.Processes()
	for _, p := range ps {
		n, _ := p.Name()
		list = append(list, models.ProcInfo{PID: p.Pid, Name: n})
	}

	// Batch CPU/MEM collection for better performance
	for i := range list {
		p, err := process.NewProcess(list[i].PID)
		if err != nil {
			continue
		}
		list[i].Cpu, _ = p.CPUPercent()
		list[i].Mem, _ = p.MemoryPercent()
	}

	_ = htmlRender.ExecuteTemplate(w, "system", list)
}
