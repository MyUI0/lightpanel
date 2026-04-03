package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"lightpanel/config"
)

type ScriptInfo struct {
	Deps  []string `json:"deps"`
	Ports []string `json:"ports"`
	Env   []string `json:"env"`
}

func analyzeScriptHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	url := r.URL.Query().Get("url")
	if url == "" {
		json.NewEncoder(w).Encode(map[string]any{"error": "missing url"})
		return
	}

	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		json.NewEncoder(w).Encode(map[string]any{"error": "invalid url scheme"})
		return
	}
	if isPrivateURL(url) {
		json.NewEncoder(w).Encode(map[string]any{"error": "private network blocked"})
		return
	}

	resp, err := http.Get(url)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]any{"error": err.Error()})
		return
	}
	defer resp.Body.Close()

	content, _ := io.ReadAll(io.LimitReader(resp.Body, 200*1024))
	script := string(content)

	scriptDir := filepath.Join(config.DataDir, "scripts")
	os.MkdirAll(scriptDir, 0755)
	fname := filepath.Base(url)
	if fname == "" || fname == "." || !strings.HasSuffix(strings.ToLower(fname), ".sh") {
		fname = "script.sh"
	}
	os.WriteFile(filepath.Join(scriptDir, fname), content, 0644)

	info := ScriptInfo{}
	seen := make(map[string]bool)

	addIfNew := func(list *[]string, val string) {
		if val != "" && !seen[val] {
			*list = append(*list, val)
			seen[val] = true
		}
	}

	for _, re := range []*regexp.Regexp{
		regexp.MustCompile(`(?:apt|apt-get|yum|dnf|apk)\s+(?:install|add)\s+(?:-y\s+)?([^\s&;#]+)`),
		regexp.MustCompile(`pip3?\s+install\s+([^\s&;#]+)`),
		regexp.MustCompile(`npm\s+(?:install|i)\s+([^\s&;#]+)`),
	} {
		for _, m := range re.FindAllStringSubmatch(script, -1) {
			addIfNew(&info.Deps, m[len(m)-1])
		}
	}

	for _, re := range []*regexp.Regexp{
		regexp.MustCompile(`(?i)(?:port|listen)\s*[=:]\s*(\d+)`),
		regexp.MustCompile(`--port\s+(\d+)`),
	} {
		for _, m := range re.FindAllStringSubmatch(script, -1) {
			addIfNew(&info.Ports, m[1])
		}
	}

	for _, m := range regexp.MustCompile(`export\s+([A-Z_]+)=`).FindAllStringSubmatch(script, -1) {
		addIfNew(&info.Env, m[1])
	}

	json.NewEncoder(w).Encode(info)
}

func scriptAnalyzePage(w http.ResponseWriter, r *http.Request) {
	_ = htmlRender.ExecuteTemplate(w, "script_analyze", nil)
}
