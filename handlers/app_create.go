package handlers

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"lightpanel/config"
	"lightpanel/models"
)

func createApp(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	name := strings.TrimSpace(r.Form.Get("name"))
	down := strings.TrimSpace(r.Form.Get("url"))
	cmd := strings.TrimSpace(r.Form.Get("cmd"))
	if name == "" || cmd == "" || strings.ContainsAny(name, `./\ `) {
		http.Redirect(w, r, "/", 302)
		return
	}

	sandbox := filepath.Join(config.AppsDir, name)
	if err := os.MkdirAll(sandbox, 0755); err != nil {
		http.Redirect(w, r, "/", 302)
		return
	}

	if down != "" {
		client := &http.Client{Timeout: 60 * time.Second}
		if resp, e := client.Get(down); e == nil {
			fname := filepath.Base(down)
			if fname == "" || fname == "." || fname == "/" {
				fname = "download"
			}
			f, err := os.Create(filepath.Join(sandbox, fname))
			if err == nil {
				_, _ = io.CopyN(f, resp.Body, config.MaxDownBytes)
				_ = f.Close()
				if strings.HasSuffix(fname, ".sh") {
					_ = os.Chmod(filepath.Join(sandbox, fname), 0755)
				}
			}
			_ = resp.Body.Close()
		}
	}

	var apps map[string]models.Project
	_ = LoadJSON(config.ConfigApps, &apps)
	apps[name] = models.Project{
		Path:    sandbox,
		Cmd:     cmd,
		Created: time.Now().Format("2006-01-02 15:04"),
	}
	_ = WriteJSON(config.ConfigApps, apps)
	http.Redirect(w, r, "/", 302)
}
