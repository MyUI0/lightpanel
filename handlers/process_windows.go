package handlers

import (
	"context"
	"github.com/shirou/gopsutil/v3/process"
	"lightpanel/models"
	"os/exec"
	"strconv"
)

// Windows is supported for local development; production targets remain Linux.
func shellCommand(ctx context.Context, command string) *exec.Cmd {
	return exec.CommandContext(ctx, "cmd.exe", "/C", command)
}
func isProcessAlive(pid int) bool {
	if pid <= 1 {
		return false
	}
	ok, _ := process.PidExists(int32(pid))
	return ok
}
func killByPid(pid int) {
	if pid > 1 {
		p, err := process.NewProcess(int32(pid))
		if err == nil {
			_ = p.Kill()
		}
	}
}
func killAppProcess(pid int) {
	if pid > 1 {
		_ = exec.Command("taskkill", "/T", "/F", "/PID", strconv.Itoa(pid)).Run()
	}
}
func processBelongsToApp(pid int, app *models.Project) bool { return managedIdentityMatches(pid, app) }
