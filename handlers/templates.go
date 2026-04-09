package handlers

import (
	"fmt"
	"html/template"
	"strings"
	"sync"
	"time"
)

var (
	htmlRender *template.Template
	renderOnce sync.Once
)

func InitTemplates() {
	renderOnce.Do(func() {
		funcMap := template.FuncMap{
			"tolower":    strings.ToLower,
			"escape":     template.HTMLEscapeString,
			"formatSize": formatSize,
			"formatTime": func(t time.Time) string { return t.Format("2006-01-02 15:04") },
			"getBgUrl":   getBgURL,
			"getLogoUrl": getLogoURL,
		}

		tmpl := template.New("").Funcs(funcMap)
		tmpl = template.Must(tmpl.Parse(
			`{{define "login"}}`+htmlLogin+`{{end}}`+
				`{{define "index"}}`+htmlIndex+`{{end}}`+
				`{{define "apps"}}`+htmlApps+`{{end}}`+
				`{{define "store"}}`+htmlStore+`{{end}}`+
				`{{define "store_params"}}`+htmlStoreParams+`{{end}}`+
				`{{define "source"}}`+htmlSource+`{{end}}`+
				`{{define "system"}}`+htmlSystem+`{{end}}`+
				`{{define "log_list"}}`+htmlLogList+`{{end}}`+
				`{{define "log"}}`+htmlLog+`{{end}}`+
				`{{define "setting"}}`+htmlSetting+`{{end}}`+
				`{{define "edit"}}`+htmlEdit+`{{end}}`+
				`{{define "downloads"}}`+htmlDownloads+`{{end}}`+
				`{{define "analyze"}}`+htmlScriptAnalyze+`{{end}}`,
		))

		htmlRender = tmpl
	})
}

func formatSize(b int64) string {
	if b <= 0 {
		return "0 B"
	}
	units := []string{"B", "KB", "MB", "GB", "TB"}
	i := 0
	for b >= 1024 && i < len(units)-1 {
		b /= 1024
		i++
	}
	return fmt.Sprintf("%.1f %s", float64(b), units[i])
}
