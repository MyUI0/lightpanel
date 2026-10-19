package handlers

import (
	"context"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"lightpanel/config"
)

var hookMu sync.Mutex

func RunHook(name string, args ...string) {
	hookDir := config.DataDir + "/hooks"
	script := filepath.Join(hookDir, name+".sh")
	if _, err := os.Stat(script); os.IsNotExist(err) {
		return
	}

	go func() {
		hookMu.Lock()
		defer hookMu.Unlock()

		shell := "sh"
		if path, err := exec.LookPath("bash"); err == nil {
			shell = path
		}
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()
		script, _ = filepath.Abs(script)
		cmd := exec.CommandContext(ctx, shell, script)
		cmd.Args = append(cmd.Args, args...)
		cmd.Dir = config.DataDir
		cmd.Stdout = nil
		cmd.Stderr = nil
		_ = cmd.Run()
	}()
}

func getCustomPages() []string {
	pagesDir := config.DataDir + "/pages"
	var pages []string
	entries, err := os.ReadDir(pagesDir)
	if err != nil {
		return nil
	}
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(strings.ToLower(e.Name()), ".html") {
			pages = append(pages, e.Name())
		}
	}
	return pages
}

func serveCustomPage(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/page/")
	pagesDir := config.DataDir + "/pages"
	if name == "" || strings.ContainsAny(name, "/\\") || !strings.HasSuffix(strings.ToLower(name), ".html") {
		http.NotFound(w, r)
		return
	}
	realPath, err := archiveTarget(pagesDir, name)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	content, err := os.ReadFile(realPath)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Security-Policy", "sandbox allow-scripts; default-src 'none'; script-src 'unsafe-inline'; style-src 'unsafe-inline'; img-src data: https:; connect-src 'none'; form-action 'none'; base-uri 'none'")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(content)
}
