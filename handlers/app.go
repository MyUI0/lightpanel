package handlers

import (
	"context"
	"lightpanel/config"
	"lightpanel/models"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

// All lifecycle changes and these maps are guarded by appOpMu.
// Generations invalidate delayed retries after stop, restart or deletion.
var generations = make(map[string]uint64)
var crashCounts = make(map[string]crashCounter)

type crashCounter struct {
	count int
	last  time.Time
}

const maxCrashRestarts = 3
const crashResetInterval = 5 * time.Minute

func checkAppStatus(app *models.Project) {
	app.Status, app.PID = "已停止", 0
	identity := readProcessIdentity(app)
	if managedIdentityMatches(identity.PID, app) {
		app.PID = identity.PID
		app.Status = "运行中"
	}
}

// Caller must hold appOpMu; launchApp never re-enters that mutex.
func launchApp(name string, app *models.Project) bool {
	checkAppStatus(app)
	if app.PID > 0 {
		return false
	}
	logFile := filepath.Join(app.Path, "run.log")
	rotateLog(logFile)
	logF, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return false
	}
	marker := filepath.Join(app.Path, ".setup_done")
	if app.SetupCmd != "" {
		if _, err := os.Stat(marker); os.IsNotExist(err) {
			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			cmd := shellCommand(ctx, app.SetupCmd)
			cmd.Dir, cmd.Stdout, cmd.Stderr = app.Path, logF, logF
			err := cmd.Run()
			cancel()
			if err != nil {
				_, _ = logF.WriteString("[setup] " + err.Error() + "\n")
				logF.Close()
				return false
			}
			if err := os.WriteFile(marker, []byte("done"), 0600); err != nil {
				logF.Close()
				return false
			}
		}
	}
	if app.Cmd == "" {
		dir := app.Path
		if app.WorkDir != "" {
			dir = app.WorkDir
		}
		app.Cmd = findBinaryInDir(dir)
		if app.Cmd == "" {
			_, _ = logF.WriteString("[start] 请填写启动命令\n")
			logF.Close()
			return false
		}
	}
	cmd := shellCommand(context.Background(), app.Cmd)
	cmd.Dir = app.Path
	if app.WorkDir != "" {
		cmd.Dir = app.WorkDir
	}
	cmd.Stdout, cmd.Stderr = logF, logF
	if err := cmd.Start(); err != nil {
		logF.WriteString(err.Error() + "\n")
		logF.Close()
		return false
	}
	pidFile := filepath.Join(app.Path, "pid.pid")
	if err := writeProcessIdentity(pidFile, cmd.Process.Pid); err != nil {
		killAppProcess(cmd.Process.Pid)
		_ = cmd.Wait()
		logF.Close()
		return false
	}
	generations[name]++
	generation := generations[name]
	app.PID, app.Status = cmd.Process.Pid, "运行中"
	started := time.Now()
	go func() {
		err := cmd.Wait()
		logF.Close()
		appOpMu.Lock()
		if generations[name] != generation {
			appOpMu.Unlock()
			return
		}
		_ = os.Remove(pidFile)
		if err == nil {
			appOpMu.Unlock()
			return
		}
		appOpMu.Unlock()
		time.Sleep(3 * time.Second)
		appOpMu.Lock()
		defer appOpMu.Unlock()
		if generations[name] != generation {
			return
		}
		var apps map[string]models.Project
		if LoadJSON(config.ConfigApps, &apps) != nil {
			return
		}
		current, ok := apps[name]
		if !ok || !current.AutoStart {
			return
		}
		checkAppStatus(&current)
		if current.PID > 0 {
			return
		}
		cc := crashCounts[name]
		if time.Since(started) >= crashResetInterval || time.Since(cc.last) >= crashResetInterval {
			cc = crashCounter{}
		}
		cc.count++
		cc.last = time.Now()
		crashCounts[name] = cc
		go notifyAppCrash(name, cc.count)
		go RunHook("app_crash", name, strconv.Itoa(cc.count))
		if cc.count > maxCrashRestarts {
			go notifyAppStopped(name)
			go RunHook("app_stopped", name)
			return
		}
		if launchApp(name, &current) {
			apps[name] = current
			_ = WriteJSON(config.ConfigApps, apps)
		}
	}()
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
	n, _ := f.ReadAt(tail, info.Size()-int64(config.MaxLogLen))
	f.Close()
	if n > 0 {
		_ = os.WriteFile(path, tail[:n], 0600)
	}
}
