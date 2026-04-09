package handlers

import "strings"

func sidebarHTML(active string) string {
	items := []struct{ path, icon, long, short string }{
		{"/", "fa-gauge", "面板首页", "首页"},
		{"/apps", "fa-layer-group", "管理应用", "应用"},
		{"/store", "fa-store", "应用商店", "商店"},
		{"/downloads", "fa-download", "下载管理", "下载"},
		{"/analyze", "fa-code", "脚本分析", "分析"},
		{"/system", "fa-chart-line", "系统监控", "监控"},
		{"/setting", "fa-gear", "系统设置", "设置"},
	}
	var nav string
	for _, it := range items {
		cls := ""
		if it.path == active || (it.path != "/" && strings.HasPrefix(active, it.path)) {
			cls = "active"
		}
		nav += `<a href="` + it.path + `" class="` + cls + `"><i class="fa-solid ` + it.icon + `"></i><span class="nav-long">` + it.long + `</span><span class="nav-short">` + it.short + `</span></a>`
	}
	for _, page := range getCustomPages() {
		pageName := strings.TrimSuffix(page, ".html")
		pagePath := "/page/" + page
		cls := ""
		if pagePath == active {
			cls = "active"
		}
		nav += `<a href="` + pagePath + `" class="` + cls + `"><i class="fa-solid fa-file"></i><span class="nav-long">` + pageName + `</span><span class="nav-short">` + pageName + `</span></a>`
	}

	logoURL := getLogoURL()
	logoHTML := `<i class="fa-solid fa-server"></i>`
	if logoURL != "" {
		logoHTML = `<img src="` + logoURL + `" alt="logo" onerror="this.style.display='none';this.parentElement.innerHTML='<i class=\\'fa-solid fa-server\\'></i>'">`
	}

	return `<div class="sidebar" id="sidebar">
<div class="logo-row">
<div class="logo-icon">` + logoHTML + `</div>
<span class="logo-text">朱雀面板</span>
</div>
<nav>` + nav + `</nav>
<div class="side-footer">
<a href="#" id="themeBtn" onclick="return false"><i class="fa-solid fa-moon"></i><span class="nav-long">切换主题</span><span class="nav-short">主题</span></a>
<a href="/logout" class="logout"><i class="fa-solid fa-right-from-bracket"></i><span class="nav-long">退出登录</span><span class="nav-short">退出</span></a>
</div>
</div>`
}

func topbarHTML(title string) string {
	return `<div class="topbar">
<div class="toggle-btn" id="toggleBtn" title="切换侧栏">&#9776;</div>
<span class="page-title">` + title + `</span>
</div>`
}
