package handlers

import (
	"encoding/json"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/process"
	"lightpanel/config"
	"lightpanel/models"
	"net/http"
	"time"
)

func apiOverview(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	apps := make(map[string]models.Project)
	if err := LoadJSON(config.ConfigApps, &apps); err != nil {
		http.Error(w, "无法读取应用配置", 500)
		return
	}
	type state struct {
		Status string `json:"status"`
		PID    int    `json:"pid"`
	}
	states := make(map[string]state)
	for name, app := range apps {
		checkAppStatus(&app)
		states[name] = state{app.Status, app.PID}
	}
	memory, diskPercent := 0, 0
	if value, err := mem.VirtualMemory(); err == nil {
		memory = int(value.UsedPercent)
	}
	if value, err := disk.Usage(config.DataDir); err == nil {
		diskPercent = int(value.UsedPercent)
	}
	pids, _ := process.Pids()
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	_ = json.NewEncoder(w).Encode(map[string]any{"cpu": getCpuPercent(), "mem": memory, "disk": diskPercent, "processes": len(pids), "apps": states, "updated_at": time.Now().Unix()})
}
