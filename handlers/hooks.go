package handlers

import (
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"lightpanel/config"
)

var hookMu sync.Mutex

func RunHook(name string, args ...string) {
	hookDir := config.DataDir + "/hooks"
	script := filepath.Join(hookDir, name+".sh")
	if _, err := os.Stat(script); os.IsNotExist(err) {
		return
	}

	hookMu.Lock()
	defer hookMu.Unlock()

	shellOnce.Do(func() { shellPath = findShell() })
	cmd := exec.Command(shellPath, script)
	cmd.Args = append(cmd.Args, args...)
	cmd.Dir = config.DataDir
	cmd.Stdout = nil
	cmd.Stderr = nil
	_ = cmd.Run()
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
	filePath := filepath.Join(pagesDir, name)

	if !strings.HasSuffix(strings.ToLower(filePath), ".html") {
		http.NotFound(w, r)
		return
	}

	realPath, err := filepath.EvalSymlinks(filePath)
	if err != nil || !strings.HasPrefix(realPath, filepath.Clean(pagesDir)) {
		http.NotFound(w, r)
		return
	}

	content, err := os.ReadFile(realPath)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(content)
}
