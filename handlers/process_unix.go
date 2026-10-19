//go:build !windows

package handlers

import (
	"context"
	"lightpanel/models"
	"os/exec"
	"syscall"
	"time"
)

func shellCommand(ctx context.Context, command string) *exec.Cmd {
	shell := "sh"
	if path, err := exec.LookPath("bash"); err == nil {
		shell = path
	}
	cmd := exec.CommandContext(ctx, shell, "-c", command)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error { return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL) }
	return cmd
}
func isProcessAlive(pid int) bool { return pid > 1 && syscall.Kill(pid, 0) == nil }
func killByPid(pid int) {
	if pid > 1 {
		_ = syscall.Kill(pid, syscall.SIGTERM)
	}
}
func killAppProcess(pid int) {
	if pid <= 1 {
		return
	}
	group, err := syscall.Getpgid(pid)
	target := pid
	if err == nil && group == pid {
		target = -pid
	}
	_ = syscall.Kill(target, syscall.SIGTERM)
	for i := 0; i < 20; i++ {
		if syscall.Kill(target, 0) != nil {
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
	_ = syscall.Kill(target, syscall.SIGKILL)
}
func processBelongsToApp(pid int, app *models.Project) bool { return managedIdentityMatches(pid, app) }
