package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/shirou/gopsutil/v3/process"
	"lightpanel/models"
	"os"
	"path/filepath"
)

type processIdentity struct {
	PID     int   `json:"pid"`
	Started int64 `json:"started"`
}

func writeProcessIdentity(path string, pid int) error {
	p, err := process.NewProcess(int32(pid))
	if err != nil {
		return err
	}
	started, err := p.CreateTime()
	if err != nil || started <= 0 {
		return fmt.Errorf("无法确认启动进程身份")
	}
	return WriteJSON(path, processIdentity{PID: pid, Started: started})
}
func readProcessIdentity(app *models.Project) processIdentity {
	var identity processIdentity
	data, err := os.ReadFile(filepath.Join(app.Path, "pid.pid"))
	if err != nil {
		return identity
	}
	if json.Unmarshal(data, &identity) != nil {
		return processIdentity{}
	}
	return identity
}
func managedIdentityMatches(pid int, app *models.Project) bool {
	identity := readProcessIdentity(app)
	if pid <= 1 || identity.PID != pid || identity.Started <= 0 {
		return false
	}
	p, err := process.NewProcess(int32(pid))
	if err != nil {
		return false
	}
	started, err := p.CreateTime()
	return err == nil && started == identity.Started && isProcessAlive(pid)
}
