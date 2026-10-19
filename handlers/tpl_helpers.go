package handlers

import (
	"html"
	"lightpanel/config"
	"net/url"
	"strings"
)

func sidebarHTML(active string) string {
	items := []struct{ path, icon, long, short string }{
		{"/", "fa-server", "应用管理", "应用"},
		{"/store", "fa-store", "应用商店", "商店"},
		{"/downloads", "fa-download", "下载管理", "下载"},
		{"/analyze", "fa-code", "脚本分析", "分析"},
		{"/system", "fa-chart-line", "系统监控", "监控"},
		{"/setting", "fa-gear", "系统设置", "设置"},
	}
	var nav string
	for index, it := range items {
		cls := ""
		if it.path == active || (it.path != "/" && strings.HasPrefix(active, it.path)) {
			cls = "active"
		}
		nav += `<a href="` + it.path + `" class="` + cls + `"><span class="nav-icon">` + string(rune('1'+index)) + `</span><span class="nav-long">` + it.long + `</span><span class="nav-short">` + it.short + `</span></a>`
	}
	for _, page := range getCustomPages() {
		pageName := strings.TrimSuffix(page, ".html")
		pagePath := "/page/" + url.PathEscape(page)
		cls := ""
		if pagePath == active {
			cls = "active"
		}
		nav += `<a href="` + pagePath + `" class="` + cls + `"><i class="fa-solid fa-file"></i><span class="nav-long">` + html.EscapeString(pageName) + `</span><span class="nav-short">` + html.EscapeString(pageName) + `</span></a>`
	}

	logoURL := getLogoURL()
	logoHTML := `L`
	if logoURL != "" {
		logoHTML = `<img src="` + html.EscapeString(logoURL) + `" alt="朱雀面板">`
	}

	return `<div class="sidebar" id="sidebar">
<div class="logo-row">
<div class="logo-icon">` + logoHTML + `</div>
<span class="logo-text">朱雀面板<small>LIGHTPANEL</small></span>
</div>
<div class="nav-caption">工作空间</div><nav aria-label="主导航">` + nav + `</nav>
<div class="side-footer">
<a href="javascript:void(0)" id="themeBtn"><i class="fa-solid fa-moon"></i><span class="nav-long">切换主题</span><span class="nav-short">主题</span></a>
<a href="/logout" class="logout"><i class="fa-solid fa-right-from-bracket"></i><span class="nav-long">退出登录</span><span class="nav-short">退出</span></a>
</div>
</div>`
}

func topbarHTML(title string) string {
	return `<header class="topbar"><button type="button" class="toggle-btn" id="toggleBtn" aria-label="切换导航" aria-controls="sidebar">☰</button><span class="page-title">` + html.EscapeString(title) + `</span><div class="topbar-meta"><span class="status-dot"></span> 控制台 <span class="version">` + html.EscapeString(config.Version) + `</span></div></header>`
}
