package handlers

import (
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"lightpanel/config"
	"lightpanel/models"
)

func checkAppStatus(app *models.Project) {
	pidFile := filepath.Join(app.Path, "pid.pid")
	app.Status = "已停止"
	app.PID = 0
	if b, e := os.ReadFile(pidFile); e == nil {
		if pid, err := strconv.Atoi(strings.TrimSpace(string(b))); err == nil && pid > 0 {
			p, e := os.FindProcess(pid)
			if e == nil && p.Signal(syscall.Signal(0)) == nil {
				app.Status = "运行中"
				app.PID = pid
			} else {
				_ = os.Remove(pidFile)
			}
		}
	}
}

func killByPid(pid int) {
	if pid <= 0 {
		return
	}
	_ = syscall.Kill(-pid, syscall.SIGTERM)
	_ = syscall.Kill(-pid, syscall.SIGKILL)
}

var (
	watchdogLock  sync.Map
	crashCounters sync.Map
)

const maxCrashRestarts = 3
const crashResetInterval = 5 * time.Minute

type crashCounter struct {
	count int
	last  time.Time
}

func launchApp(name string, app models.Project) bool {
	pidFile := filepath.Join(app.Path, "pid.pid")
	logFile := filepath.Join(app.Path, "run.log")

	if b, e := os.ReadFile(pidFile); e == nil {
		if pid, err := strconv.Atoi(strings.TrimSpace(string(b))); err == nil && pid > 0 {
			if p, e := os.FindProcess(pid); e == nil && p.Signal(syscall.Signal(0)) == nil {
				return false
			}
		}
		_ = os.Remove(pidFile)
	}

	rotateLog(logFile)

	cmd := exec.Command("bash", "-c", app.Cmd)
	cmd.Dir = app.Path
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	logF, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return false
	}
	cmd.Stdout = logF
	cmd.Stderr = logF

	if err := cmd.Start(); err != nil {
		_ = logF.Close()
		return false
	}

	_ = os.WriteFile(pidFile, []byte(strconv.Itoa(cmd.Process.Pid)), 0644)

	go func(myPid int) {
		_ = cmd.Wait()
		_ = logF.Close()

		curB, curE := os.ReadFile(pidFile)
		if curE == nil && strings.TrimSpace(string(curB)) == strconv.Itoa(myPid) {
			_ = os.Remove(pidFile)
		}

		if _, loaded := watchdogLock.LoadOrStore(name, struct{}{}); !loaded {
			defer watchdogLock.Delete(name)

			var apps map[string]models.Project
			if err := LoadJSON(config.ConfigApps, &apps); err != nil {
				return
			}
			curApp, ok := apps[name]
			if !ok || !curApp.AutoStart {
				return
			}

			val, _ := crashCounters.Load(name)
			cc, ok := val.(crashCounter)
			if !ok || time.Since(cc.last) > crashResetInterval {
				cc = crashCounter{}
			}
			cc.count++
			cc.last = time.Now()
			crashCounters.Store(name, cc)

			if cc.count > maxCrashRestarts {
				return
			}

			time.Sleep(3 * time.Second)

			if b, e := os.ReadFile(pidFile); e == nil {
				if pid, err := strconv.Atoi(strings.TrimSpace(string(b))); err == nil && pid > 0 {
					if p, e := os.FindProcess(pid); e == nil && p.Signal(syscall.Signal(0)) == nil {
						return
					}
				}
				_ = os.Remove(pidFile)
			}

			if launchApp(name, curApp) {
				crashCounters.Delete(name)
				curApp.Status = "运行中"
				apps[name] = curApp
				_ = WriteJSON(config.ConfigApps, apps)
			}
		}
	}(cmd.Process.Pid)

	return true
}

func rotateLog(path string) {
	info, err := os.Stat(path)
	if err != nil || info.Size() < 512*1024 {
		return
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return
	}
	if len(b) > config.MaxLogLen {
		b = b[len(b)-config.MaxLogLen:]
		_ = os.WriteFile(path, b, 0644)
	}
}
