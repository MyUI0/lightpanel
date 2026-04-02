package config

const (
	Version = "v2.1.5"
	Port    = "31956"

	DataDir   = "./data"
	ConfigDir = DataDir + "/config"
	AppsDir   = DataDir + "/apps"

	ConfigApps = ConfigDir + "/apps.json"
	ConfigSrc  = ConfigDir + "/sources.json"
	ConfigUsr  = ConfigDir + "/user.json"
	ConfigBg   = ConfigDir + "/bg.json"

	MaxLogLen    = 12000
	MaxDownBytes = 500 * 1024 * 1024
)
