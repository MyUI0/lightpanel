package handlers

import (
	"encoding/json"
	"html/template"
	"io"
	"net/http"
	"regexp"
	"strings"

	"time"
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

	req, err := http.NewRequestWithContext(r.Context(), "GET", url, nil)
	if err != nil {
		http.Error(w, "无效地址", 400)
		return
	}
	resp, err := remoteClient(15 * time.Second).Do(req)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]any{"error": err.Error()})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		http.Error(w, "远程服务器返回错误", 502)
		return
	}
	content, err := io.ReadAll(io.LimitReader(resp.Body, 200*1024+1))
	if err != nil || len(content) > 200*1024 {
		http.Error(w, "脚本读取失败或超过 200 KB", 400)
		return
	}
	script := string(content)

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
	_ = htmlRender.ExecuteTemplate(w, "analyze", map[string]any{
		"Sidebar": template.HTML(sidebarHTML("/analyze")),
		"Topbar":  template.HTML(topbarHTML("脚本分析")),
	})
}
