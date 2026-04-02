package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"lightpanel/config"
	"lightpanel/models"
)

func storePage(w http.ResponseWriter, r *http.Request) {
	var srcs []models.StoreSource
	_ = LoadJSON(config.ConfigSrc, &srcs)
	idx, _ := strconv.Atoi(r.URL.Query().Get("source"))
	if idx < 0 || idx >= len(srcs) {
		idx = 0
	}

	var apps []models.StoreApp
	if len(srcs) > 0 {
		client := &http.Client{Timeout: 10 * time.Second}
		if resp, e := client.Get(srcs[idx].URL); e == nil {
			defer resp.Body.Close()
			if resp.StatusCode == 200 {
				_ = json.NewDecoder(resp.Body).Decode(&apps)
			}
		}
	}
	if len(apps) == 0 {
		apps = []models.StoreApp{{
			Name:   "获取失败",
			Desc:   "网络异常或源地址错误",
			Icon:   "https://cdn-icons-png.flaticon.com/512/1164/1164100.png",
			Author: "system",
		}}
	}

	_ = htmlRender.ExecuteTemplate(w, "store", map[string]any{
		"Apps":    apps,
		"Sources": srcs,
		"Active":  idx,
	})
}

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
	var srcs []models.StoreSource
	_ = LoadJSON(config.ConfigSrc, &srcs)
	srcs = append(srcs, models.StoreSource{name, url})
	_ = WriteJSON(config.ConfigSrc, srcs)
	http.Redirect(w, r, "/source", 302)
}

func delSource(w http.ResponseWriter, r *http.Request) {
	i, _ := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/source/del/"))
	var srcs []models.StoreSource
	_ = LoadJSON(config.ConfigSrc, &srcs)
	if i >= 0 && i < len(srcs) {
		srcs = append(srcs[:i], srcs[i+1:]...)
		_ = WriteJSON(config.ConfigSrc, srcs)
	}
	http.Redirect(w, r, "/source", 302)
}
