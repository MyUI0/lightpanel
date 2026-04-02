package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"

	"lightpanel/config"
	"lightpanel/models"
)

func LoadJSON(path string, v any) error {
	b, e := os.ReadFile(path)
	if e != nil {
		return e
	}
	return json.Unmarshal(b, v)
}

func WriteJSON(path string, v any) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	tmp := path + ".tmp"
	f, e := os.Create(tmp)
	if e != nil {
		return e
	}

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	err := enc.Encode(v)
	if err == nil {
		err = f.Close()
	} else {
		f.Close()
	}

	if err != nil {
		_ = os.Remove(tmp)
		return err
	}

	return os.Rename(tmp, path)
}

func EnsureDir(path string) error {
	return os.MkdirAll(path, 0755)
}

func InitConfig() error {
	if err := EnsureDir(config.ConfigDir); err != nil {
		return err
	}
	if err := EnsureDir(config.AppsDir); err != nil {
		return err
	}

	if _, e := os.Stat(config.ConfigApps); e != nil {
		if err := WriteJSON(config.ConfigApps, map[string]models.Project{}); err != nil {
			return err
		}
	}
	if _, e := os.Stat(config.ConfigSrc); e != nil {
		if err := WriteJSON(config.ConfigSrc, []models.StoreSource{
			{"默认演示源", "https://raw.githubusercontent.com/example/test/main/apps.json"},
		}); err != nil {
			return err
		}
	}
	if _, e := os.Stat(config.ConfigUsr); e != nil {
		if err := WriteJSON(config.ConfigUsr, models.UserConfig{"admin", "admin"}); err != nil {
			return err
		}
	}
	if _, e := os.Stat(config.ConfigBg); e != nil {
		if err := WriteJSON(config.ConfigBg, models.BgConfig{Type: "gradient", URL: ""}); err != nil {
			return err
		}
	}
	return nil
}
