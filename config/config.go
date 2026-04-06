package config

import "os"

var (
	Version = "v1.0.5"
	Port    = getEnv("LIGHTPANEL_PORT", "31956")

	DataDir   = getDataDir()
	ConfigDir = DataDir + "/config"
	AppsDir   = DataDir + "/apps"

	ConfigApps = ConfigDir + "/apps.json"
	ConfigSrc  = ConfigDir + "/sources.json"
	ConfigUsr  = ConfigDir + "/user.json"
	ConfigBg   = ConfigDir + "/bg.json"
	ConfigDl   = ConfigDir + "/downloads.json"
	ConfigSys  = ConfigDir + "/system.json"

	MaxLogLen         = 102400
	MaxDownBytes      = 500 * 1024 * 1024
	MaxDownloadHistory = 10
)

type SystemInfo struct {
	Arch string `json:"arch"`
	OS   string `json:"os"`
}

func getDataDir() string {
	if dir := os.Getenv("LIGHTPANEL_DATA_DIR"); dir != "" {
		return dir
	}
	return "./data"
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
