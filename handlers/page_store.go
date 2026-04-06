package handlers

import (
	"html/template"
	"net/http"
	"strconv"
	"strings"
	"time"

	"lightpanel/config"
	"lightpanel/models"
)

func sourcePage(w http.ResponseWriter, r *http.Request) {
	var srcs []models.StoreSource
	_ = LoadJSON(config.ConfigSrc, &srcs)
	_ = htmlRender.ExecuteTemplate(w, "source", map[string]any{
		"Sources": srcs,
		"Sidebar": template.HTML(sidebarHTML("/source")),
		"Topbar":  template.HTML(topbarHTML("源管理")),
	})
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
	storeCacheLock.Lock()
	var newSrcs []models.StoreSource
	LoadJSON(config.ConfigSrc, &newSrcs)
	initStoreCache(newSrcs, len(newSrcs)-1)
	storeCacheLock.Unlock()
	http.Redirect(w, r, "/store?source="+strconv.Itoa(len(srcs)-1), 302)
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
		storeCacheLock.Lock()
		storeCache.apps = nil
		storeCache.lastInit = time.Time{}
		storeCacheLock.Unlock()
	}
	http.Redirect(w, r, "/source", 302)
}

func editSource(w http.ResponseWriter, r *http.Request) {
	i, _ := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/source/edit/"))
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
	if i >= 0 && i < len(srcs) {
		srcs[i] = models.StoreSource{Name: name, URL: url}
		_ = WriteJSON(config.ConfigSrc, srcs)
		storeCacheLock.Lock()
		storeCache.apps = nil
		storeCache.lastInit = time.Time{}
		storeCacheLock.Unlock()
	}
	http.Redirect(w, r, "/source", 302)
}
