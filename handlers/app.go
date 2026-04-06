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

	found := false

	// 1. 尝试通过 PID 文件检查
	if b, e := os.ReadFile(pidFile); e == nil {
		if pid, err := strconv.Atoi(strings.TrimSpace(string(b))); err == nil && pid > 0 {
			if isProcessAlive(pid) {
				app.Status = "运行中"
				app.PID = pid
				found = true
			}
		}
	}

	// 2. 如果 PID 无效，尝试通过进程名查找（支持守护进程模式）
	if !found && app.Cmd != "" {
		parts := strings.Fields(app.Cmd)
		if len(parts) > 0 {
			binName := filepath.Base(parts[0])
			if binName != "sh" && binName != "bash" && binName != "./" && binName != "" {
				if pid := findProcessByName(binName); pid > 0 {
					app.Status = "运行中"
					app.PID = pid
					found = true
					_ = os.WriteFile(pidFile, []byte(strconv.Itoa(pid)), 0644)
				}
			}
		}
	}

	if !found {
		_ = os.Remove(pidFile)
	}
}

func isProcessAlive(pid int) bool {
	p, e := os.FindProcess(pid)
	return e == nil && p.Signal(syscall.Signal(0)) == nil
}

func findProcessByName(name string) int {
	processCache.mu.RLock()
	if cached, ok := processCache.data[name]; ok && time.Since(cached.time) < processCacheTTL {
		processCache.mu.RUnlock()
		return cached.pid
	}
	processCache.mu.RUnlock()

	entries, _ := os.ReadDir("/proc")
	for _, entry := range entries {
		if _, err := strconv.Atoi(entry.Name()); err != nil {
			continue
		}
		cmdLinePath := filepath.Join("/proc", entry.Name(), "cmdline")
		data, err := os.ReadFile(cmdLinePath)
		if err != nil {
			continue
		}
		args := strings.Split(string(data), "\x00")
		if len(args) > 0 {
			if filepath.Base(args[0]) == name {
				pid, _ := strconv.Atoi(entry.Name())
				processCache.mu.Lock()
				processCache.data[name] = struct{ pid int; time time.Time }{pid: pid, time: time.Now()}
				processCache.mu.Unlock()
				return pid
			}
		}
	}
	return 0
}

func killByPid(pid int) {
	if pid <= 0 {
		return
	}
	_ = syscall.Kill(pid, syscall.SIGTERM)
	time.Sleep(100 * time.Millisecond)
	_ = syscall.Kill(pid, syscall.SIGKILL)
}

var (
	watchdogLock     sync.Mutex
	watchdogRunning  map[string]bool
	crashCounters    sync.Map
	crashCountersMu  sync.Mutex
	processCache     struct {
		mu   sync.RWMutex
		data map[string]struct{ pid int; time time.Time }
	}
	processCacheTTL = 30 * time.Second
)

const maxCrashRestarts = 3
const crashResetInterval = 5 * time.Minute
const maxCrashCounters = 500

func init() {
	watchdogRunning = make(map[string]bool)
	processCache.data = make(map[string]struct{ pid int; time time.Time })

	go func() {
		ticker := time.NewTicker(10 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			now := time.Now()
			crashCountersMu.Lock()
			crashCounters.Range(func(key, value any) bool {
				if cc, ok := value.(crashCounter); ok {
					if now.Sub(cc.last) > crashResetInterval*2 {
						crashCounters.Delete(key)
					}
				}
				return true
			})
			var crashCountersLen int
			crashCounters.Range(func(key, value any) bool {
				crashCountersLen++
				return true
			})
			if crashCountersLen > maxCrashCounters {
				crashCountersMu.Unlock()
				var toDelete []string
				crashCounters.Range(func(key, value any) bool {
					if cc, ok := value.(crashCounter); ok {
						if now.Sub(cc.last) > crashResetInterval {
							toDelete = append(toDelete, key.(string))
						}
					}
					return true
				})
				for _, k := range toDelete {
					crashCounters.Delete(k)
				}
				crashCountersMu.Unlock()
			} else {
				crashCountersMu.Unlock()
			}
		}
	}()
}

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
			done := make(chan error, 1)
			go func() {
				done <- setupCmd.Run()
			}()
			select {
			case <-time.After(60 * time.Second):
				if setupCmd.Process != nil {
					_ = setupCmd.Process.Kill()
				}
				_, _ = setupLog.Write([]byte("[setup] 首次命令执行超时\n"))
				return false
			case err := <-done:
				if err != nil {
					_, _ = setupLog.Write([]byte("[setup] 首次命令执行失败: " + err.Error() + "\n"))
					return false
				}
			}
			_, _ = setupLog.Write([]byte("[setup] 首次命令执行完成\n"))
			_ = os.WriteFile(setupMarker, []byte("done"), 0644)
		}
	}

	// 如果 Cmd 为空，在沙盒目录中自动查找可执行文件
	if app.Cmd == "" {
		if bin := findBinaryInDir(app.Path); bin != "" {
			app.Cmd = bin
			appOpMu.Lock()
			var apps map[string]models.Project
			if err := LoadJSON(config.ConfigApps, &apps); err == nil {
				if curApp, ok := apps[name]; ok {
					curApp.Cmd = bin
					apps[name] = curApp
					_ = WriteJSON(config.ConfigApps, apps)
				}
			}
			appOpMu.Unlock()
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
			// 进程已退出，但不删除 PID 文件，等待 cmd.Wait() 判断退出码
		}
	}()

	go func(myPid int) {
		defer func() { recover() }()
		err := cmd.Wait()
		_ = logF.Close()

		curB, curE := os.ReadFile(pidFile)
		if curE == nil && strings.TrimSpace(string(curB)) == strconv.Itoa(myPid) {
			if err == nil {
				return
			}
			if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 0 {
				return
			}
			_ = os.Remove(pidFile)
		}

		watchdogLock.Lock()
		if watchdogRunning[name] {
			watchdogLock.Unlock()
			return
		}
		watchdogRunning[name] = true
		watchdogLock.Unlock()

		defer func() {
			watchdogLock.Lock()
			delete(watchdogRunning, name)
			watchdogLock.Unlock()
		}()

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
		var apps map[string]models.Project
		if err := LoadJSON(config.ConfigApps, &apps); err != nil {
			appOpMu.Unlock()
			return
		}
		curApp, ok := apps[name]
		if !ok || !curApp.AutoStart {
			appOpMu.Unlock()
			watchdogLock.Lock()
			delete(watchdogRunning, name)
			watchdogLock.Unlock()
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

		if cc.count > maxCrashRestarts {
			appOpMu.Unlock()
			notifyAppStopped(name)
			go RunHook("app_stopped", name)
			return
		}
		appOpMu.Unlock()

		if launchApp(name, &curApp) {
			appOpMu.Lock()
			crashCounters.Delete(name)
			curApp.Status = "运行中"
			apps[name] = curApp
			_ = WriteJSON(config.ConfigApps, apps)
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
