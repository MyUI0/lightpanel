package config

import "os"

var (
	Version = "v3.0.1"
	Port    = getEnv("LIGHTPANEL_PORT", "31956")

	DataDir   = getDataDir()
	ConfigDir = DataDir + "/config"
	AppsDir   = DataDir + "/apps"

	ConfigApps = ConfigDir + "/apps.json"
	ConfigSrc  = ConfigDir + "/sources.json"
	ConfigUsr  = ConfigDir + "/user.json"
	ConfigBg   = ConfigDir + "/bg.json"

	MaxLogLen    = 12000
	MaxDownBytes = 500 * 1024 * 1024
)

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
