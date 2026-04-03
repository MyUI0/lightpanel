package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"lightpanel/config"
	"lightpanel/models"
)

func sourcePage(w http.ResponseWriter, r *http.Request) {
	var srcs []models.StoreSource
	_ = LoadJSON(config.ConfigSrc, &srcs)
	_ = htmlRender.ExecuteTemplate(w, "source", srcs)
}

func addSource(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	name := strings.TrimSpace(r.Form.Get("name"))
	url := strings.TrimSpace(r.Form.Get("url"))
	if name == "" || url == "" {
		http.Redirect(w, r, "/source", 302)
		return
	}
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		http.Redirect(w, r, "/source", 302)
		return
	}
	if isPrivateURL(url) {
		http.Redirect(w, r, "/source", 302)
		return
	}
	appOpMu.Lock()
	defer appOpMu.Unlock()
	var srcs []models.StoreSource
	_ = LoadJSON(config.ConfigSrc, &srcs)
	srcs = append(srcs, models.StoreSource{Name: name, URL: url})
	_ = WriteJSON(config.ConfigSrc, srcs)
	http.Redirect(w, r, "/source", 302)
}

func delSource(w http.ResponseWriter, r *http.Request) {
	i, _ := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/source/del/"))
	appOpMu.Lock()
	defer appOpMu.Unlock()
	var srcs []models.StoreSource
	_ = LoadJSON(config.ConfigSrc, &srcs)
	if i >= 0 && i < len(srcs) {
		srcs = append(srcs[:i], srcs[i+1:]...)
		_ = WriteJSON(config.ConfigSrc, srcs)
	}
	http.Redirect(w, r, "/source", 302)
}
