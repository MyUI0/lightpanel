package handlers

import (
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

var (
	shellPath string
	shellOnce sync.Once
)

func findShell() string {
	if _, err := exec.LookPath("bash"); err == nil {
		return "bash"
	}
	return "sh"
}

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

func launchApp(name string, app *models.Project) bool {
	pidFile := filepath.Join(app.Path, "pid.pid")
	logFile := filepath.Join(app.Path, "run.log")
	setupMarker := filepath.Join(app.Path, ".setup_done")

	if b, e := os.ReadFile(pidFile); e == nil {
		if pid, err := strconv.Atoi(strings.TrimSpace(string(b))); err == nil && pid > 0 {
			if p, e := os.FindProcess(pid); e == nil && p.Signal(syscall.Signal(0)) == nil {
				return false
			}
		}
		_ = os.Remove(pidFile)
	}

	rotateLog(logFile)

	shellOnce.Do(func() { shellPath = findShell() })

	if app.SetupCmd != "" {
		if _, err := os.Stat(setupMarker); os.IsNotExist(err) {
			setupLog, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
			if err != nil {
				return false
			}
			defer setupLog.Close()
			_, _ = setupLog.Write([]byte("\n[setup] 执行首次命令: " + app.SetupCmd + "\n"))
			setupCmd := exec.Command(shellPath, "-c", app.SetupCmd)
			setupCmd.Dir = app.Path
			setupCmd.Stdout = setupLog
			setupCmd.Stderr = setupLog
			err = setupCmd.Run()
			if err != nil {
				_, _ = setupLog.Write([]byte("[setup] 首次命令执行失败: " + err.Error() + "\n"))
				return false
			}
			_, _ = setupLog.Write([]byte("[setup] 首次命令执行完成\n"))
			_ = os.WriteFile(setupMarker, []byte("done"), 0644)
		}
	}

	// 如果 Cmd 为空，在沙盒目录中自动查找可执行文件
	if app.Cmd == "" {
		if bin := findBinaryInDir(app.Path); bin != "" {
			app.Cmd = bin
			var apps map[string]models.Project
			if err := LoadJSON(config.ConfigApps, &apps); err == nil {
				if _, ok := apps[name]; ok {
					apps[name] = *app
					_ = WriteJSON(config.ConfigApps, apps)
				}
			}
		}
	}

	if app.Cmd == "" {
		return false
	}

	cmd := exec.Command(shellPath, "-c", app.Cmd)
	cmd.Dir = app.Path
	if app.WorkDir != "" {
		cmd.Dir = app.WorkDir
	}
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

	pidData := []byte(strconv.Itoa(cmd.Process.Pid))
	tmpPid := pidFile + ".tmp"
	_ = os.WriteFile(tmpPid, pidData, 0644)
	_ = os.Rename(tmpPid, pidFile)

	go func() {
		defer func() { recover() }()
		time.Sleep(500 * time.Millisecond)
		if p, e := os.FindProcess(cmd.Process.Pid); e != nil || p.Signal(syscall.Signal(0)) != nil {
			if b, e := os.ReadFile(pidFile); e == nil && strings.TrimSpace(string(b)) == strconv.Itoa(cmd.Process.Pid) {
				_ = os.Remove(pidFile)
			}
		}
	}()

	go func(myPid int) {
		defer func() { recover() }()
		_ = cmd.Wait()
		_ = logF.Close()

		curB, curE := os.ReadFile(pidFile)
		if curE == nil && strings.TrimSpace(string(curB)) == strconv.Itoa(myPid) {
			_ = os.Remove(pidFile)
		}

		if _, loaded := watchdogLock.LoadOrStore(name, struct{}{}); !loaded {
			defer watchdogLock.Delete(name)

			appOpMu.Lock()
			var apps map[string]models.Project
			if err := LoadJSON(config.ConfigApps, &apps); err != nil {
				appOpMu.Unlock()
				return
			}
			curApp, ok := apps[name]
			if !ok || !curApp.AutoStart {
				appOpMu.Unlock()
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

			notifyAppCrash(name, cc.count)
			go RunHook("app_crash", name, strconv.Itoa(cc.count))

			if cc.count >= maxCrashRestarts {
				appOpMu.Unlock()
				notifyAppStopped(name)
				go RunHook("app_stopped", name)
				return
			}
			appOpMu.Unlock()

			time.Sleep(3 * time.Second)

			if b, e := os.ReadFile(pidFile); e == nil {
				if pid, err := strconv.Atoi(strings.TrimSpace(string(b))); err == nil && pid > 0 {
					if p, e := os.FindProcess(pid); e == nil && p.Signal(syscall.Signal(0)) == nil {
						return
					}
				}
				_ = os.Remove(pidFile)
			}

			appOpMu.Lock()
			if launchApp(name, &curApp) {
				crashCounters.Delete(name)
				curApp.Status = "运行中"
				apps[name] = curApp
				_ = WriteJSON(config.ConfigApps, apps)
			}
			appOpMu.Unlock()
		}
	}(cmd.Process.Pid)

	go RunHook("app_start", name, strconv.Itoa(cmd.Process.Pid))

	return true
}

func rotateLog(path string) {
	info, err := os.Stat(path)
	if err != nil || info.Size() < 512*1024 {
		return
	}
	f, err := os.Open(path)
	if err != nil {
		return
	}
	tail := make([]byte, config.MaxLogLen)
	offset := info.Size() - int64(config.MaxLogLen)
	if offset < 0 {
		offset = 0
	}
	n, _ := f.ReadAt(tail, offset)
	_ = f.Close()
	if n > 0 {
		tmpPath := path + ".tmp"
		_ = os.WriteFile(tmpPath, tail[:n], 0644)
		_ = os.Rename(tmpPath, path)
	}
}
