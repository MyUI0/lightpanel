package handlers

import (
	"encoding/json"
	"os"
	"path/filepath"

	"lightpanel/config"
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

	_ = os.Chmod(tmp, 0600)
	return os.Rename(tmp, path)
}

func EnsureDir(path string) error {
	return os.MkdirAll(path, 0755)
}

func getLogoURL() string {
	type LogoConfig struct {
		URL string `json:"url"`
	}
	var logo LogoConfig
	_ = LoadJSON(config.ConfigDir+"/logo.json", &logo)
	return logo.URL
}
