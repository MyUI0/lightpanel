package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type BarkConfig struct {
	Enabled bool   `json:"enabled"`
	Device  string `json:"device"`
	Group   string `json:"group"`
}

func loadBarkConfig() BarkConfig {
	var cfg BarkConfig
	_ = LoadJSON(config.ConfigDir+"/bark.json", &cfg)
	return cfg
}

func sendBark(title, body string) {
	cfg := loadBarkConfig()
	if !cfg.Enabled || cfg.Device == "" {
		return
	}

	deviceKey := cfg.Device
	if !strings.HasPrefix(deviceKey, "http") {
		deviceKey = "https://api.day.app/" + deviceKey
	}

	payload := map[string]any{
		"title": title,
		"body":  body,
		"group": cfg.Group,
	}
	if cfg.Group == "" {
		delete(payload, "group")
	}

	data, _ := json.Marshal(payload)
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(deviceKey, "application/json", bytes.NewReader(data))
	if err != nil {
		return
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)
}

func notifyAppCrash(name string, crashCount int) {
	sendBark(
		fmt.Sprintf("⚠️ 应用崩溃: %s", name),
		fmt.Sprintf("第 %d 次崩溃（最多 %d 次自动重启）", crashCount, maxCrashRestarts),
	)
}

func notifyAppStopped(name string) {
	sendBark("🛑 应用已停止", fmt.Sprintf("应用 %s 已停止运行", name))
}
