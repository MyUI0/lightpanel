package handlers

import (
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"strings"
	"sync"
	"time"
)

//go:embed web/templates/*.html web/static/*
var webFiles embed.FS

func staticHandler() http.Handler {
	sub, _ := fs.Sub(webFiles, "web/static")
	return http.StripPrefix("/static/", http.FileServer(http.FS(sub)))
}

var (
	htmlRender *template.Template
	renderOnce sync.Once
)

func InitTemplates() {
	renderOnce.Do(func() {
		funcMap := template.FuncMap{
			"tolower": strings.ToLower,
			"initial": func(s string) string {
				for _, r := range s {
					return strings.ToUpper(string(r))
				}
				return "L"
			},
			"escape":     template.HTMLEscapeString,
			"formatSize": formatSize,
			"formatTime": func(t time.Time) string { return t.Format("2006-01-02 15:04") },
			"getBgUrl":   getBgURL,
			"getLogoUrl": getLogoURL,
		}

		tmpl := template.Must(template.New("").Funcs(funcMap).ParseFS(webFiles, "web/templates/*.html"))

		htmlRender = tmpl
	})
}

func formatSize(b int64) string {
	if b <= 0 {
		return "0 B"
	}
	units := []string{"B", "KB", "MB", "GB", "TB"}
	value := float64(b)
	i := 0
	for value >= 1024 && i < len(units)-1 {
		value /= 1024
		i++
	}
	return fmt.Sprintf("%.1f %s", value, units[i])
}
