package handlers

import "strings"

const layoutCSS = `@import url('https://fonts.googleapis.com/css2?family=Inter:wght@300;400;500;600;700&display=swap');
:root{--accent:#6366f1;--bg:#0f0f23;--bg2:rgba(15,15,35,0.35);--card:rgba(255,255,255,0.03);--card-h:rgba(255,255,255,0.06);--text:#e2e8f0;--text2:rgba(255,255,255,0.45);--border:rgba(255,255,255,0.06);--input-bg:rgba(255,255,255,0.04);--input-b:rgba(255,255,255,0.08);--input-t:#e2e8f0;--input-ph:rgba(255,255,255,0.3);--sidebar-w:200px;--sidebar-cw:70px}
[data-theme="light"]{--bg:#e8eaed;--bg2:rgba(255,255,255,0.55);--card:rgba(255,255,255,0.6);--card-h:rgba(255,255,255,0.8);--text:#1a1a2e;--text2:rgba(0,0,0,0.55);--border:rgba(0,0,0,0.1);--input-bg:rgba(255,255,255,0.7);--input-b:rgba(0,0,0,0.12);--input-t:#1a1a2e;--input-ph:rgba(0,0,0,0.35)}
*{margin:0;padding:0;box-sizing:border-box;font-family:'Inter',system-ui,sans-serif}
body{background:var(--bg);min-height:100vh;color:var(--text);display:flex}
.bg-layer{position:fixed;inset:0;z-index:0;background-size:cover;background-position:center;background-repeat:no-repeat}
.bg-layer::before{content:'';position:absolute;inset:0;background:radial-gradient(ellipse at 50% 30%,rgba(99,102,241,0.12) 0%,transparent 60%),var(--bg);opacity:0.9}
.sidebar{width:var(--sidebar-w);min-height:100vh;background:var(--bg2);backdrop-filter:blur(30px);-webkit-backdrop-filter:blur(30px);border-right:1px solid var(--border);display:flex;flex-direction:column;position:fixed;left:0;top:0;bottom:0;z-index:100;transition:width 0.25s ease;overflow:hidden}
.sidebar.collapsed{width:var(--sidebar-cw)}
.sidebar .logo-row{display:flex;align-items:center;gap:0.6rem;padding:1.2rem 1rem;border-bottom:1px solid var(--border);justify-content:center}
.sidebar .logo-icon{width:36px;height:36px;border-radius:50%;display:flex;align-items:center;justify-content:center;flex-shrink:0;position:relative;overflow:hidden;border:2px solid rgba(99,102,241,0.4);box-shadow:0 0 15px rgba(99,102,241,0.3)}
.sidebar .logo-icon img{width:100%;height:100%;object-fit:cover;border-radius:50%}
.sidebar .logo-icon i{color:#fff;font-size:1rem}
.sidebar .logo-text{font-size:0.9rem;font-weight:700;color:var(--text);white-space:nowrap;transition:opacity 0.2s}
.sidebar.collapsed .logo-text{opacity:0;pointer-events:none;display:none}
.sidebar nav{flex:1;padding:0.5rem 0.4rem;display:flex;flex-direction:column;gap:2px}
.sidebar nav a{display:flex;align-items:center;gap:0.7rem;padding:0.65rem 0.7rem;border-radius:10px;color:var(--text2);text-decoration:none;font-size:0.8rem;font-weight:500;transition:all 0.15s;white-space:nowrap;justify-content:flex-start}
.sidebar nav a:hover{background:var(--card-h);color:var(--text)}
.sidebar nav a.active{background:rgba(99,102,241,0.15);color:#818cf8}
.sidebar nav a i{width:18px;text-align:center;font-size:0.9rem;flex-shrink:0}
.sidebar nav a .nav-long{transition:opacity 0.2s}
.sidebar nav a .nav-short{display:none;transition:opacity 0.2s}
.sidebar.collapsed nav{display:block;padding:0.5rem 0}
.sidebar.collapsed nav a{display:flex;justify-content:center;align-items:center;padding:0.35rem 0.3rem;gap:0;width:80%;margin:0 auto}
.sidebar.collapsed nav a i{display:none}
.sidebar.collapsed nav a .nav-long{display:none}
.sidebar.collapsed nav a .nav-short{display:block;font-size:0.82rem;font-weight:600;text-align:center;line-height:1}
.sidebar .side-footer{padding:0.6rem 0.4rem;border-top:1px solid var(--border);display:flex;flex-direction:column;gap:2px}
.sidebar .side-footer a{display:flex;align-items:center;gap:0.7rem;padding:0.6rem 0.7rem;border-radius:10px;color:var(--text2);text-decoration:none;font-size:0.8rem;transition:all 0.15s;white-space:nowrap;justify-content:flex-start}
.sidebar .side-footer a:hover{background:var(--card-h);color:var(--text)}
.sidebar .side-footer a i{width:18px;text-align:center;font-size:0.85rem;flex-shrink:0}
.sidebar .side-footer a .nav-long{transition:opacity 0.2s}
.sidebar .side-footer a .nav-short{display:none}
.sidebar.collapsed .side-footer{display:block}
.sidebar.collapsed .side-footer a{display:flex;justify-content:center;align-items:center;padding:0.35rem 0.3rem;gap:0;width:80%;margin:0 auto}
.sidebar.collapsed .side-footer a i{display:none}
.sidebar.collapsed .side-footer a .nav-long{display:none}
.sidebar.collapsed .side-footer a .nav-short{display:block;font-size:0.82rem;font-weight:600;text-align:center;line-height:1}
.sidebar .side-footer .logout{color:#f87171}
.main{margin-left:var(--sidebar-w);flex:1;min-height:100vh;display:flex;flex-direction:column;transition:margin-left 0.25s ease;position:relative;z-index:1}
.sidebar.collapsed~.main{margin-left:var(--sidebar-cw)}
.topbar{display:flex;align-items:center;gap:0.8rem;padding:0.7rem 1.5rem;border-bottom:1px solid var(--border);background:rgba(15,15,35,0.4);backdrop-filter:blur(20px);-webkit-backdrop-filter:blur(20px);position:sticky;top:0;z-index:50}
[data-theme="light"] .topbar{background:rgba(255,255,255,0.6)}
.topbar .toggle-btn{width:30px;height:30px;border-radius:50%;border:1px solid var(--border);background:var(--card);color:var(--text2);display:flex;align-items:center;justify-content:center;cursor:pointer;transition:all 0.15s;flex-shrink:0;font-size:1.1rem;line-height:1}
.topbar .toggle-btn:hover{background:var(--card-h);color:var(--text)}
.topbar .page-title{font-size:0.95rem;font-weight:600;color:var(--text)}
.topbar .spacer{flex:1}
.content{flex:1;padding:1.2rem 1.5rem}
.card{background:var(--card);backdrop-filter:blur(30px);-webkit-backdrop-filter:blur(30px);border:1px solid var(--border);border-radius:14px;padding:1rem;transition:all 0.2s}
.card:hover{background:var(--card-h)}
.glass{background:var(--card);backdrop-filter:blur(30px);-webkit-backdrop-filter:blur(30px);border:1px solid var(--border);border-radius:14px}
.btn{display:inline-flex;align-items:center;justify-content:center;gap:0.4rem;padding:0.5rem 1rem;border-radius:9px;font-size:0.75rem;font-weight:500;transition:all 0.15s;cursor:pointer;border:none;text-decoration:none}
.btn-primary{background:linear-gradient(135deg,#6366f1,#7c3aed);color:#fff}
.btn-primary:hover{box-shadow:0 3px 12px rgba(99,102,241,0.3)}
.btn-ghost{background:var(--card);color:var(--text2);border:1px solid var(--border)}
.btn-ghost:hover{background:var(--card-h);color:var(--text)}
.btn-danger{background:linear-gradient(135deg,#ef4444,#dc2626);color:#fff}
.btn-success{background:linear-gradient(135deg,#10b981,#059669);color:#fff}
.btn-warning{background:linear-gradient(135deg,#f59e0b,#d97706);color:#fff}
.btn-sm{padding:0.4rem 0.8rem;font-size:0.75rem;border-radius:7px}
.input{width:100%;padding:0.6rem 0.8rem;border-radius:9px;background:var(--input-bg);border:1px solid var(--input-b);color:var(--input-t);font-size:0.8rem;outline:none;transition:all 0.2s}
.input:focus{border-color:#6366f1;box-shadow:0 0 0 3px rgba(99,102,241,0.15)}
.input::placeholder{color:var(--input-ph)}
.alert{padding:0.6rem 0.8rem;border-radius:9px;font-size:0.78rem;margin-bottom:0.8rem}
.alert-error{background:rgba(239,68,68,0.1);color:#f87171;border:1px solid rgba(239,68,68,0.2)}
.alert-success{background:rgba(16,185,129,0.1);color:#34d399;border:1px solid rgba(16,185,129,0.2)}
.badge{display:inline-flex;align-items:center;gap:0.3rem;padding:0.2rem 0.6rem;border-radius:9999px;font-size:0.68rem;font-weight:500}
.badge-running{background:rgba(16,185,129,0.1);color:#34d399;border:1px solid rgba(16,185,129,0.2)}
.badge-stopped{background:rgba(239,68,68,0.1);color:#f87171;border:1px solid rgba(239,68,68,0.2)}
.stat-card{text-align:center;padding:0.8rem}
.stat-value{font-size:1.3rem;font-weight:700;color:var(--text)}
.stat-label{font-size:0.7rem;color:var(--text2);margin-top:0.2rem}
.progress-bar{width:100%;height:6px;background:rgba(255,255,255,0.06);border-radius:3px;overflow:hidden}
.progress-fill{height:100%;background:linear-gradient(90deg,#6366f1,#a855f7);border-radius:3px;transition:width 0.3s}
.check-label{display:flex;align-items:center;gap:0.4rem;padding:0.6rem 0.8rem;border-radius:9px;background:var(--card);border:1px solid var(--border);cursor:pointer;transition:all 0.15s}
.check-label:has(input:checked){background:rgba(16,185,129,0.08);border-color:rgba(16,185,129,0.25)}
.check-label input{display:none}
.check-box{width:18px;height:18px;border-radius:5px;border:2px solid var(--input-b);display:flex;align-items:center;justify-content:center;transition:all 0.15s}
.check-label:has(input:checked) .check-box{background:#10b981;border-color:#10b981}
.check-label:has(input:checked) .check-box i{color:#fff;font-size:0.6rem}
.fail-card{background:rgba(239,68,68,0.08);border:1px solid rgba(239,68,68,0.2);border-radius:12px;padding:0.8rem;margin-bottom:0.8rem}
.fail-card h4{color:#f87171;font-size:0.82rem;margin-bottom:0.4rem}
.fail-card pre{font-size:0.7rem;color:var(--text2);white-space:pre-wrap;word-break:break-all;max-height:120px;overflow-y:auto}
.proc-wrap{overflow-x:auto}
.proc-table{width:100%;border-collapse:collapse}
.proc-table th,.proc-table td{padding:0.6rem 0.8rem;text-align:left;font-size:0.78rem;border-bottom:1px solid var(--border)}
.proc-table th{color:var(--text2);font-weight:500}
.proc-table td{color:var(--text)}`

const layoutJS = `<script>
(function(){
var csrfCookie=document.cookie.match(/(?:^|; )lp_csrf=([^;]*)/);
var csrfToken=csrfCookie?decodeURIComponent(csrfCookie[1]):'';
if(csrfToken){
var forms=document.querySelectorAll('form[method="POST"], form[action]');
for(var i=0;i<forms.length;i++){
var f=forms[i];
var act=f.getAttribute('action')||'';
if(act.indexOf('/login/auth')===0||act.indexOf('/logout')===0)continue;
var inp=document.createElement('input');
inp.type='hidden';inp.name='csrf_token';inp.value=csrfToken;
f.appendChild(inp);
}
}
var sb=document.getElementById('sidebar');
var tb=document.getElementById('toggleBtn');
if(!sb||!tb)return;
var s=localStorage.getItem('lp_sidebar');
if(s==='collapsed'){sb.classList.add('collapsed');}
tb.addEventListener('click',function(){
sb.classList.toggle('collapsed');
localStorage.setItem('lp_sidebar',sb.classList.contains('collapsed')?'collapsed':'expanded');
});
var b=document.getElementById('themeBtn');
if(b){
var t=localStorage.getItem('lp_theme');
if(t==='light'){applyTheme('light');}
b.addEventListener('click',function(){
var c=document.documentElement.getAttribute('data-theme');
if(c==='light'){applyTheme('dark');}
else{applyTheme('light');}
});
}
function applyTheme(mode){
if(mode==='light'){
document.documentElement.setAttribute('data-theme','light');
document.documentElement.style.setProperty('--bg','#e8eaed');
document.documentElement.style.setProperty('--bg2','rgba(255,255,255,0.55)');
document.documentElement.style.setProperty('--card','rgba(255,255,255,0.6)');
document.documentElement.style.setProperty('--card-h','rgba(255,255,255,0.8)');
document.documentElement.style.setProperty('--text','#1a1a2e');
document.documentElement.style.setProperty('--text2','rgba(0,0,0,0.55)');
document.documentElement.style.setProperty('--border','rgba(0,0,0,0.1)');
document.documentElement.style.setProperty('--input-bg','rgba(255,255,255,0.7)');
document.documentElement.style.setProperty('--input-b','rgba(0,0,0,0.12)');
document.documentElement.style.setProperty('--input-t','#1a1a2e');
document.documentElement.style.setProperty('--input-ph','rgba(0,0,0,0.35)');
if(b){var ic=b.querySelector('i');if(ic){ic.className='fa-solid fa-sun';}}
localStorage.setItem('lp_theme','light');
}else{
document.documentElement.removeAttribute('data-theme');
document.documentElement.style.removeProperty('--bg');
document.documentElement.style.removeProperty('--bg2');
document.documentElement.style.removeProperty('--card');
document.documentElement.style.removeProperty('--card-h');
document.documentElement.style.removeProperty('--text');
document.documentElement.style.removeProperty('--text2');
document.documentElement.style.removeProperty('--border');
document.documentElement.style.removeProperty('--input-bg');
document.documentElement.style.removeProperty('--input-b');
document.documentElement.style.removeProperty('--input-t');
document.documentElement.style.removeProperty('--input-ph');
if(b){var ic=b.querySelector('i');if(ic){ic.className='fa-solid fa-moon';}}
localStorage.setItem('lp_theme','dark');
}
}
})();
</script>`

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
	logoInner := `<i class="fa-solid fa-server"></i>`
	if logoURL := getLogoURL(); logoURL != "" {
		logoInner = `<img src="` + logoURL + `" alt="logo" onerror="this.style.display='none';this.parentElement.innerHTML='<i class=\\'fa-solid fa-server\\'></i>'">`
	}
	return `<div class="sidebar" id="sidebar">
<div class="logo-row">
<div class="logo-icon">{{if getLogoUrl}}<img src="{{getLogoUrl}}" alt="logo" onerror="this.style.display='none';this.parentElement.innerHTML='<i class=\\'fa-solid fa-server\\'></i>'">{{else}}<i class="fa-solid fa-server"></i>{{end}}</div>
<span class="logo-text">LightPanel</span>
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

var htmlLogin = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width,initial-scale=1.0">
<title>登录 - LightPanel</title>
<style>
*{margin:0;padding:0;box-sizing:border-box;font-family:'Inter',system-ui,sans-serif}
body{min-height:100vh;display:flex;align-items:center;justify-content:center;background:#0f0f23;position:relative;overflow:hidden}
body::before{content:'';position:absolute;inset:0;background:radial-gradient(ellipse at 30% 50%,rgba(99,102,241,0.15),transparent 60%),radial-gradient(ellipse at 70% 50%,rgba(168,85,247,0.1),transparent 50%)}
.card{position:relative;z-index:1;background:rgba(255,255,255,0.04);backdrop-filter:blur(30px);border:1px solid rgba(255,255,255,0.08);border-radius:20px;padding:2.5rem;width:360px}
.logo{width:48px;height:48px;background:linear-gradient(135deg,#6366f1,#a855f7);border-radius:50%;display:flex;align-items:center;justify-content:center;margin:0 auto 1.2rem;box-shadow:0 0 15px rgba(99,102,241,0.4);border:2px solid rgba(255,255,255,0.2)}
.logo i{color:#fff;font-size:1.3rem}
h1{text-align:center;font-size:1.3rem;font-weight:700;color:#fff;margin-bottom:0.3rem}
.sub{text-align:center;font-size:0.8rem;color:rgba(255,255,255,0.4);margin-bottom:1.8rem}
.field{margin-bottom:0.8rem}
.field label{display:block;font-size:0.75rem;font-weight:600;color:rgba(255,255,255,0.5);margin-bottom:0.3rem}
.field input{width:100%;padding:0.65rem 0.85rem;border-radius:10px;border:1px solid rgba(255,255,255,0.1);background:rgba(255,255,255,0.05);color:#fff;font-size:0.85rem;outline:none;transition:all 0.2s}
.field input:focus{border-color:#6366f1;box-shadow:0 0 0 3px rgba(99,102,241,0.15)}
.btn{width:100%;padding:0.75rem;border-radius:10px;border:none;background:linear-gradient(135deg,#6366f1,#7c3aed);color:#fff;font-size:0.9rem;font-weight:600;cursor:pointer;transition:all 0.2s;margin-top:0.5rem}
.btn:hover{box-shadow:0 6px 20px rgba(99,102,241,0.35)}
.btn:disabled{opacity:0.5;cursor:not-allowed}
.err{background:rgba(239,68,68,0.1);color:#f87171;padding:0.6rem 0.8rem;border-radius:8px;font-size:0.78rem;margin-bottom:0.8rem;text-align:center;border:1px solid rgba(239,68,68,0.2)}
.warn{background:rgba(245,158,11,0.1);color:#fbbf24;padding:0.6rem 0.8rem;border-radius:8px;font-size:0.78rem;margin-bottom:0.8rem;text-align:center;border:1px solid rgba(245,158,11,0.2)}
.ft{text-align:center;font-size:0.65rem;color:rgba(255,255,255,0.25);margin-top:1.2rem}
</style>
<link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.5.1/css/all.min.css">
</head>
<body>
<div class="card">
  <div class="logo"><i class="fa-solid fa-server"></i></div>
  <h1>LightPanel</h1>
  <p class="sub">服务器管理面板</p>
  {{if eq .Err "1"}}
  <div class="err"><i class="fa-solid fa-circle-exclamation" style="margin-right:0.3rem;"></i>用户名或密码错误</div>
  {{end}}
  {{if eq .Err "locked"}}
  <div class="err"><i class="fa-solid fa-lock" style="margin-right:0.3rem;"></i>登录失败次数过多，请等待 <span id="lockTimer">{{.LockTime}}</span> 秒后重试</div>
  {{end}}
  <form action="/login/auth" method="post">
    <div class="field"><label>用户名</label><input name="username" placeholder="请输入用户名" required autocomplete="username"></div>
    <div class="field"><label>密码</label><input name="password" type="password" placeholder="请输入密码" required autocomplete="current-password"></div>
    <button class="btn" type="submit" id="loginBtn"><i class="fa-solid fa-right-to-bracket" style="margin-right:0.4rem;"></i>登录</button>
  </form>
  <div class="ft">LightPanel · 轻量高效</div>
</div>
{{if eq .Err "locked"}}
<script>
(function(){
var el=document.getElementById('lockTimer');
var btn=document.getElementById('loginBtn');
if(!el||!btn)return;
var s=parseInt(el.textContent)||0;
btn.disabled=true;
var t=setInterval(function(){
s--;
if(s<=0){clearInterval(t);btn.disabled=false;el.textContent='0';}
else{el.textContent=s;}
},1000);
})();
</script>
{{end}}
</body>
</html>`

var htmlIndex = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width,initial-scale=1.0">
<title>LightPanel</title>
<style>
` + layoutCSS + `
.app-item{display:flex;align-items:center;gap:0.8rem}
.app-name{font-weight:600;color:var(--text);font-size:0.85rem}
.app-cmd{font-size:0.7rem;color:var(--text2);margin-top:0.15rem}
</style>
</head>
<body>
<div class="bg-layer"{{if .BgUrl}} style="background-image:url('{{.BgUrl}}')"{{end}}></div>
` + sidebarHTML("/") + `
<div class="main">
` + topbarHTML("应用管理") + `
<div class="content">
{{if .FirstLogin}}
<div style="background:rgba(245,158,11,0.1);border:1px solid rgba(245,158,11,0.25);border-radius:12px;padding:0.8rem 1rem;margin-bottom:0.8rem;display:flex;align-items:center;gap:0.6rem">
<i class="fa-solid fa-shield-halved" style="font-size:1.2rem;color:#fbbf24;flex-shrink:0"></i>
<div style="flex:1"><span style="font-size:0.82rem;color:#fbbf24;font-weight:600">首次登录提示：</span><span style="font-size:0.78rem;color:var(--text2)">为了您的安全，请立即修改默认密码（admin/admin）。</span></div>
<a href="/setting" class="btn btn-warning btn-sm" style="flex-shrink:0"><i class="fa-solid fa-key"></i>修改密码</a>
</div>
{{end}}
{{if .FailInfo}}
<div class="fail-card">
<h4><i class="fa-solid fa-triangle-exclamation" style="margin-right:0.4rem;"></i>应用启动失败: {{.FailInfo.Name}}{{if gt .FailInfo.Count 1}} (及其他应用){{end}}</h4>
{{if .FailInfo.Deps}}<div style="display:flex;gap:0.4rem;flex-wrap:wrap;margin-top:0.3rem">{{range .FailInfo.Deps}}<span style="background:rgba(245,158,11,0.1);color:#fbbf24;padding:0.15rem 0.5rem;border-radius:5px;font-size:0.68rem"><i class="fa-solid fa-circle-exclamation" style="margin-right:0.2rem;"></i>{{.}}</span>{{end}}</div>{{end}}
<pre>{{escape .FailInfo.Log}}</pre>
</div>
{{end}}
{{if .CreateErr}}
<div class="fail-card"><h4><i class="fa-solid fa-circle-xmark" style="margin-right:0.4rem;"></i>创建失败: {{.CreateErr}}</h4></div>
{{end}}
<div id="activeTaskBanner" style="display:none;background:rgba(99,102,241,0.08);border:1px solid rgba(99,102,241,0.2);border-radius:10px;padding:0.7rem 0.8rem;margin-bottom:0.8rem">
<div style="font-size:0.8rem;color:#a5b4fc;margin-bottom:0.4rem"><i class="fa-solid fa-spinner fa-spin" style="margin-right:0.3rem"></i><span id="activeTaskMsg">正在创建...</span></div>
<div class="progress-bar"><div class="progress-fill" id="activeTaskFill" style="width:0%"></div></div>
</div>
<div style="display:grid;grid-template-columns:repeat(4,1fr);gap:0.6rem;margin-bottom:1rem">
<div class="card stat-card"><div class="stat-value">{{.Cpu}}%</div><div class="stat-label"><i class="fa-solid fa-microchip" style="margin-right:0.2rem"></i>CPU</div></div>
<div class="card stat-card"><div class="stat-value">{{.Mem}}%</div><div class="stat-label"><i class="fa-solid fa-memory" style="margin-right:0.2rem"></i>内存</div></div>
<div class="card stat-card"><div class="stat-value">{{.Disk}}%</div><div class="stat-label"><i class="fa-solid fa-hard-drive" style="margin-right:0.2rem"></i>磁盘</div></div>
<div class="card stat-card"><div class="stat-value">{{.ProcNum}}</div><div class="stat-label"><i class="fa-solid fa-list" style="margin-right:0.2rem"></i>进程</div></div>
</div>
<div style="font-size:0.7rem;color:var(--text2);margin-bottom:1rem;text-align:center"><i class="fa-solid fa-clock" style="margin-right:0.2rem"></i>运行时间: {{.Uptime}}</div>
<div class="glass" style="padding:1rem;margin-bottom:1rem">
<div style="display:flex;gap:0.5rem;margin-bottom:1rem">
<button type="button" class="btn btn-primary btn-sm" id="tabDownload" onclick="switchTab('download')">下载安装</button>
<button type="button" class="btn btn-ghost btn-sm" id="tabManual" onclick="switchTab('manual')">手动添加</button>
</div>
<form id="createForm" style="display:grid;gap:0.5rem">
<div id="downloadPanel">
<div style="display:grid;grid-template-columns:1fr 1fr;gap:0.5rem">
<input name="name" placeholder="应用名称" required class="input" id="appName">
<input name="cmd" placeholder="启动命令（留空自动检测）" class="input" id="appCmd">
</div>
<input name="url" placeholder="下载地址（可选）" class="input" id="appUrl">
<input name="setup_cmd" placeholder="首次运行命令（可选）" class="input" id="appSetup">
<div style="display:grid;grid-template-columns:1fr 1fr;gap:0.5rem">
<label class="check-label" style="justify-content:center"><input type="checkbox" name="auto_extract" id="autoExtract"><div class="check-box"><i class="fa-solid fa-check"></i></div><span style="color:var(--text2);font-size:0.72rem">自动解压</span></label>
<label class="check-label" style="justify-content:center"><input type="checkbox" name="make_exec" id="makeExec"><div class="check-box"><i class="fa-solid fa-check"></i></div><span style="color:var(--text2);font-size:0.72rem">赋予权限</span></label>
</div>
</div>
<div id="manualPanel" style="display:none">
<div style="display:grid;grid-template-columns:1fr 1fr;gap:0.5rem">
<input name="manual_name" placeholder="应用名称" class="input" id="manualName">
<input name="manual_path" placeholder="应用目录路径" class="input" id="manualPath">
</div>
<input name="manual_cmd" placeholder="启动命令" class="input" id="manualCmd">
<input name="manual_workdir" placeholder="工作目录（可选）" class="input" id="manualWorkDir">
<input name="manual_url" placeholder="网页地址（可选）" class="input" id="manualUrl">
<label class="check-label" style="justify-content:center"><input type="checkbox" name="manual_auto" id="manualAuto"><div class="check-box"><i class="fa-solid fa-check"></i></div><span style="color:var(--text2);font-size:0.72rem">开机自启</span></label>
</div>
<button class="btn btn-primary" style="width:100%" type="submit" id="createBtn"><i class="fa-solid fa-plus"></i>创建应用</button>
</form>
<div id="createProgress" style="display:none;margin-top:0.5rem">
<div class="progress-bar" style="height:5px;margin-bottom:0.3rem"><div class="progress-fill" id="progressFill" style="width:0%"></div></div>
<p style="font-size:0.68rem;color:var(--text2);text-align:center" id="progressText">准备中...</p>
</div>
</div>
<div id="appListHeader" style="display:flex;align-items:center;gap:0.5rem;margin-bottom:0.8rem;flex-wrap:wrap">
<div style="flex:1;min-width:200px">
<input type="text" id="appSearch" placeholder="搜索应用名称或命令..." class="input" style="padding:0.5rem 0.7rem;font-size:0.78rem">
</div>
<div style="display:flex;gap:0.3rem;align-items:center">
<button class="btn btn-ghost btn-sm" id="selectAllBtn" title="全选"><i class="fa-solid fa-check-double"></i></button>
<button class="btn btn-success btn-sm" id="batchStartBtn" title="批量启动" style="display:none"><i class="fa-solid fa-play"></i></button>
<button class="btn btn-warning btn-sm" id="batchStopBtn" title="批量停止" style="display:none"><i class="fa-solid fa-stop"></i></button>
<button class="btn btn-danger btn-sm" id="batchDeleteBtn" title="批量删除" style="display:none"><i class="fa-solid fa-trash"></i></button>
<span id="selCount" style="font-size:0.7rem;color:var(--text2);display:none">已选 0</span>
</div>
</div>
{{range $name, $app := .Apps}}
<div class="card app-item" style="margin-bottom:0.5rem" data-name="{{tolower $name}}" data-cmd="{{tolower $app.Cmd}}">
<label style="display:flex;align-items:center;gap:0.5rem;flex:1;min-width:0;cursor:pointer">
<input type="checkbox" class="app-cb" data-name="{{$name}}" style="accent-color:#6366f1;flex-shrink:0">
<div style="flex:1;min-width:0">
<div style="display:flex;align-items:center;gap:0.5rem;flex-wrap:wrap">
<span class="app-name">{{$name}}</span>
{{if eq $app.Status "运行中"}}<span class="badge badge-running"><span style="width:5px;height:5px;background:#34d399;border-radius:50%"></span>运行中</span>{{else}}<span class="badge badge-stopped"><span style="width:5px;height:5px;background:#f87171;border-radius:50%"></span>已停止</span>{{end}}
{{if $app.AutoStart}}<span class="badge" style="background:rgba(59,130,246,0.1);color:#60a5fa;border:1px solid rgba(59,130,246,0.2)"><i class="fa-solid fa-rotate" style="margin-right:0.2rem"></i>自启</span>{{end}}
{{if $app.Version}}<span class="badge" style="background:rgba(99,102,241,0.1);color:#a5b4fc;border:1px solid rgba(99,102,241,0.2)">v{{$app.Version}}</span>{{end}}
{{if index $.Updates $name}}<span class="badge" style="background:rgba(245,158,11,0.2);color:#fbbf24;border:1px solid rgba(245,158,11,0.3)"><i class="fa-solid fa-arrow-up" style="margin-right:0.2rem"></i>有更新 v{{index $.Updates $name}}</span>{{end}}
</div>
<div class="app-cmd"><i class="fa-solid fa-terminal" style="margin-right:0.2rem"></i>{{$app.Cmd}}</div>
</div>
</label>
<div style="display:flex;gap:0.3rem;flex-wrap:wrap;flex-shrink:0">
{{if eq $app.Status "运行中"}}
<form action="/stop/{{$name}}" method="post"><button class="btn btn-warning btn-sm"><i class="fa-solid fa-stop"></i></button></form>
<form action="/restart/{{$name}}" method="post"><button class="btn btn-primary btn-sm"><i class="fa-solid fa-rotate"></i></button></form>
{{else}}
<form action="/start/{{$name}}" method="post"><button class="btn btn-success btn-sm"><i class="fa-solid fa-play"></i></button></form>
{{end}}
<a href="/edit/{{$name}}" class="btn btn-ghost btn-sm"><i class="fa-solid fa-pen"></i></a>
<a href="/log/{{$name}}" class="btn btn-ghost btn-sm"><i class="fa-solid fa-file-lines"></i></a>
{{if $app.URL}}<a href="{{$app.URL}}" target="_blank" class="btn btn-ghost btn-sm" title="打开网页"><i class="fa-solid fa-globe"></i></a>{{end}}
<form action="/toggle-auto/{{$name}}" method="post"><button class="btn btn-ghost btn-sm" title="切换自启"><i class="fa-solid fa-rotate"></i></button></form>
<form action="/delete/{{$name}}" method="post" onsubmit="return confirm('确定删除 {{$name}}？')"><button class="btn btn-danger btn-sm"><i class="fa-solid fa-trash"></i></button></form>
</div>
</div>
{{end}}
{{if eq (len .Apps) 0}}
<div class="glass" style="padding:2rem;text-align:center"><i class="fa-solid fa-inbox" style="font-size:2rem;color:rgba(255,255,255,0.1);margin-bottom:0.5rem"></i><p style="color:var(--text2);font-size:0.8rem">暂无应用</p></div>
{{end}}
</div>
</div>
` + layoutJS + `
<script>
(function(){
var form=document.getElementById('createForm');
var prog=document.getElementById('createProgress');
var fill=document.getElementById('progressFill');
var text=document.getElementById('progressText');
var btn=document.getElementById('createBtn');
var banner=document.getElementById('activeTaskBanner');
var bannerMsg=document.getElementById('activeTaskMsg');
var bannerFill=document.getElementById('activeTaskFill');
var downloadPanel=document.getElementById('downloadPanel');
var manualPanel=document.getElementById('manualPanel');
var tabDownload=document.getElementById('tabDownload');
var tabManual=document.getElementById('tabManual');
window.switchTab=function(tab){
if(tab==='download'){
downloadPanel.style.display='block';
manualPanel.style.display='none';
tabDownload.className='btn btn-primary btn-sm';
tabManual.className='btn btn-ghost btn-sm';
}else{
downloadPanel.style.display='none';
manualPanel.style.display='block';
tabDownload.className='btn btn-ghost btn-sm';
tabManual.className='btn btn-primary btn-sm';
}
}
if(!form)return;
var pollId=null;
var activeTaskId=localStorage.getItem('lp_createTask');
var activeTaskErr=localStorage.getItem('lp_createErr');
function showBanner(){if(banner)banner.style.display='block'}
function hideBanner(){if(banner)banner.style.display='none';localStorage.removeItem('lp_createTask');localStorage.removeItem('lp_createErr')}
function pollProgress(id){
fetch('/create/progress/'+id).then(function(r){return r.json()}).then(function(t){
if(!t||!t.status)return;
if(t.status==='creating'||t.status==='downloading'){
var pct=t.progress||0;
if(fill)fill.style.width=pct+'%';
if(text)text.textContent=t.message||'处理中... '+pct+'%';
if(bannerMsg)bannerMsg.textContent=t.message||'正在创建...';
if(bannerFill)bannerFill.style.width=pct+'%';
showBanner();pollId=setTimeout(function(){pollProgress(id)},500);
}else if(t.status==='completed'){
if(fill)fill.style.width='100%';if(text)text.textContent='创建完成！';
if(bannerFill)bannerFill.style.width='100%';if(bannerMsg)bannerMsg.textContent='创建完成！';
showBanner();clearTimeout(pollId);hideBanner();setTimeout(function(){location.reload()},600);
}else if(t.status==='error'){
if(text){text.textContent='错误: '+t.message;text.style.color='#f87171'}
if(bannerMsg)bannerMsg.textContent='创建失败: '+t.message;
if(bannerFill)bannerFill.style.width='100%';
if(banner)banner.style.borderColor='rgba(239,68,68,0.3)';
showBanner();btn.disabled=false;btn.innerHTML='<i class="fa-solid fa-plus"></i>创建应用';
clearTimeout(pollId);localStorage.setItem('lp_createErr','1');localStorage.removeItem('lp_createTask');
}else if(t.status==='not_found'){hideBanner()}
}).catch(function(){pollId=setTimeout(function(){pollProgress(id)},1000)});
}
if(activeTaskErr){
if(banner){banner.style.display='block';banner.style.borderColor='rgba(239,68,68,0.3)';bannerMsg.textContent='上次创建失败';bannerFill.style.width='100%'}
localStorage.removeItem('lp_createErr');
}
if(activeTaskId){pollProgress(activeTaskId)}
form.addEventListener('submit',function(e){
e.preventDefault();
var isManual=manualPanel.style.display==='block';
var nameEl=isManual?document.getElementById('manualName'):document.getElementById('appName');
if(!nameEl)return;
var name=nameEl.value.trim();
if(!name){nameEl.focus();return}
localStorage.removeItem('lp_createErr');
btn.disabled=true;btn.innerHTML='<i class="fa-solid fa-spinner fa-spin"></i>处理中...';
prog.style.display='block';fill.style.width='10%';text.textContent=isManual?'添加应用...':'准备创建...';
if(isManual){
var mName=document.getElementById('manualName').value.trim();
var mPath=document.getElementById('manualPath').value.trim();
var mCmd=document.getElementById('manualCmd').value.trim();
var mWorkDir=document.getElementById('manualWorkDir').value.trim();
var mUrl=document.getElementById('manualUrl').value.trim();
var mAuto=document.getElementById('manualAuto').checked;
if(!mPath||!mCmd){text.textContent='路径和命令必填';text.style.color='#f87171';btn.disabled=false;btn.innerHTML='<i class="fa-solid fa-plus"></i>创建应用';return}
var fd=new FormData();
fd.append('name',mName);fd.append('path',mPath);fd.append('cmd',mCmd);
if(mWorkDir)fd.append('work_dir',mWorkDir);
if(mUrl)fd.append('url',mUrl);
if(mAuto)fd.append('auto','on');
fetch('/create/manual',{method:'POST',body:fd}).then(function(r){return r.json()}).then(function(data){
if(data.ok){text.textContent='添加成功！';fill.style.width='100%';setTimeout(function(){location.reload()},500)}
else{text.textContent=data.error||'添加失败';text.style.color='#f87171';btn.disabled=false;btn.innerHTML='<i class="fa-solid fa-plus"></i>创建应用'}
}).catch(function(){text.textContent='网络错误';btn.disabled=false;btn.innerHTML='<i class="fa-solid fa-plus"></i>创建应用'});
return;
}
var fd=new FormData(form);
fetch('/create',{method:'POST',body:fd}).then(function(r){
var ct=r.headers.get('content-type')||'';
if(ct.indexOf('json')===-1){text.textContent='请求失败';btn.disabled=false;btn.innerHTML='<i class="fa-solid fa-plus"></i>创建应用';return}
return r.json()}).then(function(data){
if(!data)return;
if(data.ok&&data.task){localStorage.setItem('lp_createTask',data.task);pollProgress(data.task);setTimeout(function(){window.location.replace(data.redirect||'/')},300)}
else if(data.error){text.textContent=data.error;text.style.color='#f87171';if(bannerMsg)bannerMsg.textContent='创建失败: '+data.error;if(banner)banner.style.borderColor='rgba(239,68,68,0.3)';showBanner();localStorage.setItem('lp_createErr','1');btn.disabled=false;btn.innerHTML='<i class="fa-solid fa-plus"></i>创建应用'}
}).catch(function(){text.textContent='网络错误';btn.disabled=false;btn.innerHTML='<i class="fa-solid fa-plus"></i>创建应用'});
});
var searchInput=document.getElementById('appSearch');
if(searchInput){
searchInput.addEventListener('input',function(){
var q=this.value.toLowerCase();
var items=document.querySelectorAll('.app-item');
var vis=0;
for(var i=0;i<items.length;i++){
var n=items[i].getAttribute('data-name')||'';
var c=items[i].getAttribute('data-cmd')||'';
if(n.indexOf(q)>=0||c.indexOf(q)>=0){items[i].style.display='';vis++;}
else{items[i].style.display='none';}
}
});
}
var selAll=document.getElementById('selectAllBtn');
if(selAll){
selAll.addEventListener('click',function(){
var cbs=document.querySelectorAll('.app-cb');
var allChecked=true;
for(var i=0;i<cbs.length;i++){if(!cbs[i].checked){allChecked=false;break;}}
for(var i=0;i<cbs.length;i++){cbs[i].checked=!allChecked;}
updateSelCount();
});
}
function updateSelCount(){
var cbs=document.querySelectorAll('.app-cb:checked');
var cnt=cbs.length;
var el=document.getElementById('selCount');
var sb=document.getElementById('batchStartBtn');
var st=document.getElementById('batchStopBtn');
var sd=document.getElementById('batchDeleteBtn');
if(el){el.textContent='已选 '+cnt;el.style.display=cnt>0?'':'none';}
if(sb)sb.style.display=cnt>0?'':'none';
if(st)st.style.display=cnt>0?'':'none';
if(sd)sd.style.display=cnt>0?'':'none';
}
document.addEventListener('change',function(e){if(e.target&&e.target.classList.contains('app-cb'))updateSelCount();});
function batchAction(action,confirmMsg){
var cbs=document.querySelectorAll('.app-cb:checked');
if(cbs.length===0)return;
var names=[];
for(var i=0;i<cbs.length;i++){names.push(cbs[i].getAttribute('data-name'));}
if(!confirm(confirmMsg+names.join(', ')))return;
var csrfEl=document.querySelector('input[name="csrf_token"]');
for(var i=0;i<names.length;i++){
var f=document.createElement('form');
f.method='POST';f.action='/'+action+'/'+names[i];
if(csrfEl){var inp=document.createElement('input');inp.type='hidden';inp.name='csrf_token';inp.value=csrfEl.value;f.appendChild(inp);}
document.body.appendChild(f);f.submit();
}
}
var bs=document.getElementById('batchStartBtn');
if(bs)bs.addEventListener('click',function(){batchAction('start','确定启动: ')});
var bst=document.getElementById('batchStopBtn');
if(bst)bst.addEventListener('click',function(){batchAction('stop','确定停止: ')});
var bd=document.getElementById('batchDeleteBtn');
if(bd)bd.addEventListener('click',function(){batchAction('delete','确定删除: ')});
})();
</script>
</body>
</html>`

var htmlStore = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width,initial-scale=1.0">
<title>应用商店 - LightPanel</title>
<style>
` + layoutCSS + `
.store-item{display:flex;align-items:center;gap:0.8rem}
.store-icon{width:44px;height:44px;border-radius:10px;object-fit:cover;border:1px solid var(--border)}
.source-tabs{display:flex;gap:0.4rem;margin-bottom:1rem;flex-wrap:wrap}
.source-tab{padding:0.4rem 0.8rem;border-radius:8px;font-size:0.75rem;text-decoration:none;transition:all 0.15s}
.source-tab.active{background:linear-gradient(135deg,#6366f1,#7c3aed);color:#fff}
.source-tab:not(.active){background:var(--card);color:var(--text2);border:1px solid var(--border)}
</style>
</head>
<body>
<div class="bg-layer"{{if getBgUrl}} style="background-image:url('{{getBgUrl}}')"{{end}}></div>
` + sidebarHTML("/store") + `
<div class="main">
` + topbarHTML("应用商店") + `
<div class="content">
<div class="source-tabs">
{{range $i,$s := .Sources}}
{{if eq $i $.Active}}<span class="source-tab active">{{$s.Name}}</span>{{else}}<a href="/store?source={{$i}}" class="source-tab">{{$s.Name}}</a>{{end}}
{{end}}
<a href="/source" class="source-tab" style="border-style:dashed"><i class="fa-solid fa-plus"></i> 添加源</a>
{{if .StoreErr}}<a href="/source" class="source-tab" style="color:#f87171"><i class="fa-solid fa-triangle-exclamation"></i> 源配置</a>{{end}}
</div>
{{if .StoreErr}}<div class="alert alert-{{if eq .StoreErrType "network"}}error{{else if eq .StoreErrType "http"}}warning{{else}}error{{end}}" style="margin-bottom:1rem"><i class="fa-solid fa-circle-exclamation"></i> {{.StoreErr}}</div>{{end}}
<div style="margin-bottom:1rem">
<input type="text" id="storeSearch" placeholder="搜索应用..." class="input" style="width:100%" oninput="filterStore()">
</div>
<div style="display:grid;gap:0.5rem" id="storeList">
{{range $i,$a := .Apps}}
<div class="card store-item" data-name="{{$a.Name}}" data-desc="{{$a.Desc}}">
<img src="{{$a.Icon}}" class="store-icon" referrerpolicy="no-referrer" onerror="this.style.display='none'">
<div style="flex:1;min-width:0">
<h3 style="font-size:0.9rem;font-weight:600;color:var(--text)">{{$a.Name}}{{if $a.Version}}<span style="font-size:0.65rem;background:var(--card);color:var(--text2);padding:0.1rem 0.3rem;margin-left:0.3rem;border-radius:4px">v{{$a.Version}}</span>{{end}}</h3>
<p style="font-size:0.72rem;color:var(--text2);margin-top:0.15rem">{{$a.Desc}}</p>
<p style="font-size:0.65rem;color:var(--text2);margin-top:0.2rem"><i class="fa-solid fa-user" style="margin-right:0.2rem"></i>{{$a.Author}}</p>
</div>
<form action="/install/{{$i}}?source={{$.Active}}" method="post" class="install-form">
<button class="btn btn-success"><i class="fa-solid fa-download"></i>部署</button>
</form>
</div>
{{end}}
</div>
</div>
</div>
` + layoutJS + `
<script>
function filterStore(){
var kw=document.getElementById('storeSearch').value.toLowerCase();
document.querySelectorAll('.store-item').forEach(function(el){var name=el.getAttribute('data-name')||'';var desc=el.getAttribute('data-desc')||'';el.style.display=(kw&&!name.includes(kw)&&!desc.includes(kw))?'none':'flex'})}
document.querySelectorAll('.install-form').forEach(function(f){f.addEventListener('submit',function(e){var b=f.querySelector('button');if(b&&b.disabled){e.preventDefault();return}})});
</script>
</body>
</html>`

var htmlDownloads = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width,initial-scale=1.0">
<title>下载管理 - LightPanel</title>
<style>
` + layoutCSS + `
.badge-downloading{background:rgba(59,130,246,0.1);color:#60a5fa;border:1px solid rgba(59,130,246,0.2)}
.badge-paused{background:rgba(245,158,11,0.1);color:#fbbf24;border:1px solid rgba(245,158,11,0.2)}
.badge-completed{background:rgba(16,185,129,0.1);color:#34d399;border:1px solid rgba(16,185,129,0.2)}
.badge-error{background:rgba(239,68,68,0.1);color:#f87171;border:1px solid rgba(239,68,68,0.2)}
</style>
</head>
<body>
<div class="bg-layer"{{if getBgUrl}} style="background-image:url('{{getBgUrl}}')"{{end}}></div>
` + sidebarHTML("/downloads") + `
<div class="main">
` + topbarHTML("下载管理") + `
<div class="content">
<div id="taskList">
{{range .}}
<div class="card" id="task-{{.ID}}" style="margin-bottom:0.5rem">
<div style="display:flex;justify-content:space-between;align-items:center;margin-bottom:0.3rem">
<div><span style="font-size:0.88rem;font-weight:600;color:var(--text)">{{.Name}}</span><span class="badge badge-{{.Status}}" style="margin-left:0.4rem">{{if eq .Status "downloading"}}下载中{{end}}{{if eq .Status "paused"}}已暂停{{end}}{{if eq .Status "completed"}}已完成{{end}}{{if eq .Status "error"}}失败{{end}}</span></div>
<div style="display:flex;gap:0.3rem">
{{if eq .Status "downloading"}}<button class="btn btn-warning btn-sm" onclick="dlAction('{{.ID}}','pause')"><i class="fa-solid fa-pause"></i></button>{{end}}
{{if eq .Status "paused"}}<button class="btn btn-success btn-sm" onclick="dlAction('{{.ID}}','resume')"><i class="fa-solid fa-play"></i></button>{{end}}
{{if eq .Status "completed"}}<button class="btn btn-sm" style="background:linear-gradient(135deg,#3b82f6,#2563eb);color:#fff" onclick="dlAction('{{.ID}}','install')"><i class="fa-solid fa-rocket"></i></button>{{end}}
<button class="btn btn-danger btn-sm" onclick="dlAction('{{.ID}}','delete')"><i class="fa-solid fa-trash"></i></button>
</div>
</div>
<div style="font-size:0.68rem;color:var(--text2);word-break:break-all;margin-bottom:0.3rem">{{.URL}}</div>
<div class="progress-bar"><div class="progress-fill" style="width:{{.Progress}}%"></div></div>
<div style="font-size:0.68rem;color:var(--text2);display:flex;justify-content:space-between"><span class="progress-pct">{{.Progress}}%</span><span class="progress-size">{{formatSize .Downloaded}} / {{formatSize .Size}}</span></div>
</div>
{{end}}
{{if eq (len .) 0}}
<div class="glass" style="padding:2rem;text-align:center"><i class="fa-solid fa-inbox" style="font-size:2rem;color:rgba(255,255,255,0.1);margin-bottom:0.5rem"></i><p style="color:var(--text2);font-size:0.8rem">暂无下载任务</p></div>
{{end}}
</div>
</div>
</div>
` + layoutJS + `
<script>
function formatSize(b){if(b<=0)return '0 B';var u=['B','KB','MB','GB','TB'];var i=0;while(b>=1024&&i<u.length-1){b/=1024;i++}return b.toFixed(1)+' '+u[i]}
function dlAction(id,act){fetch('/dl/action/'+id+'/'+act,{method:'POST'}).then(function(){location.reload()})}
setInterval(function(){fetch('/dl/api').then(function(r){return r.json()}).then(function(tasks){tasks.forEach(function(t){var card=document.getElementById('task-'+t.id);if(!card)return;var fill=card.querySelector('.progress-fill');if(fill)fill.style.width=t.progress+'%';var pct=card.querySelector('.progress-pct');if(pct)pct.textContent=t.progress+'%';var sz=card.querySelector('.progress-size');if(sz&&t.size>0)sz.textContent=formatSize(t.downloaded)+' / '+formatSize(t.size)})})},1000);
</script>
</body>
</html>`

var htmlEdit = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width,initial-scale=1.0">
<title>编辑应用 - LightPanel</title>
<style>
` + layoutCSS + `
</style>
</head>
<body>
<div class="bg-layer"{{if getBgUrl}} style="background-image:url('{{getBgUrl}}')"{{end}}></div>
` + sidebarHTML("/edit/") + `
<div class="main">
` + topbarHTML("编辑应用") + `
<div class="content">
{{if eq .Err "running"}}<div class="alert alert-error"><i class="fa-solid fa-triangle-exclamation" style="margin-right:0.3rem"></i>请先停止应用</div>{{end}}
{{if eq .Err "exists"}}<div class="alert alert-error"><i class="fa-solid fa-triangle-exclamation" style="margin-right:0.3rem"></i>名称已存在</div>{{end}}
{{if eq .Err "rename"}}<div class="alert alert-error"><i class="fa-solid fa-triangle-exclamation" style="margin-right:0.3rem"></i>重命名失败</div>{{end}}
{{if eq .Err "save"}}<div class="alert alert-error"><i class="fa-solid fa-triangle-exclamation" style="margin-right:0.3rem"></i>保存失败</div>{{end}}
{{if eq .Msg "1"}}<div class="alert alert-success"><i class="fa-solid fa-check" style="margin-right:0.3rem"></i>已保存</div>{{end}}
{{if .InstallNote}}<div class="alert" style="background:rgba(99,102,241,0.08);color:#a5b4fc;border:1px solid rgba(99,102,241,0.2)"><i class="fa-solid fa-circle-info" style="margin-right:0.3rem"></i>安装检测: {{.InstallNote}} <a href="/detect/{{.Name}}" style="color:#818cf8;margin-left:0.3rem">[重新检测]</a></div>{{end}}
<div class="glass" style="padding:1rem">
<div style="display:flex;align-items:center;justify-content:space-between;margin-bottom:1rem">
<h3 style="font-size:0.95rem;font-weight:600;color:var(--text)">{{.Name}}</h3>
{{if eq .Status "运行中"}}<span class="badge badge-running"><span style="width:5px;height:5px;background:#34d399;border-radius:50%"></span>运行中</span>{{else}}<span class="badge badge-stopped"><span style="width:5px;height:5px;background:#f87171;border-radius:50%"></span>已停止</span>{{end}}
</div>
<form id="editForm" action="/edit/{{.Name}}" method="post" style="display:grid;gap:0.5rem">
<div><label style="font-size:0.7rem;color:var(--text2);margin-bottom:0.2rem;display:block">应用名称</label><input name="name" value="{{.Name}}" required class="input"></div>
<div><label style="font-size:0.7rem;color:var(--text2);margin-bottom:0.2rem;display:block">沙盒路径</label><input name="path" value="{{.Path}}" placeholder="留空不修改" class="input"></div>
<div><label style="font-size:0.7rem;color:var(--text2);margin-bottom:0.2rem;display:block">工作目录</label><input name="work_dir" value="{{.WorkDir}}" placeholder="留空则使用沙盒路径" class="input"></div>
<div><label style="font-size:0.7rem;color:var(--text2);margin-bottom:0.2rem;display:block">首次运行命令（仅执行一次）</label><input name="setup_cmd" value="{{.SetupCmd}}" placeholder="留空则不执行" class="input"></div>
<div><label style="font-size:0.7rem;color:var(--text2);margin-bottom:0.2rem;display:block">启动命令</label><input name="cmd" value="{{.Cmd}}" required class="input"></div>
<div><label style="font-size:0.7rem;color:var(--text2);margin-bottom:0.2rem;display:block">网页地址</label><input name="url" value="{{.URL}}" placeholder="如 http://127.0.0.1:8080" class="input"></div>
<label class="check-label"><input type="checkbox" name="auto"{{if .Auto}} checked{{end}}><div class="check-box"><i class="fa-solid fa-check"></i></div><span style="color:var(--text2);font-size:0.78rem">开机自启（崩溃自动重启）</span></label>
<div style="display:grid;grid-template-columns:1fr 1fr;gap:0.5rem;margin-top:0.5rem">
<button class="btn btn-primary" style="width:100%"><i class="fa-solid fa-save"></i>保存</button>
<a href="/" class="btn btn-ghost" style="width:100%"><i class="fa-solid fa-xmark"></i>取消</a>
</div>
</form>
<div style="display:grid;grid-template-columns:1fr 1fr;gap:0.5rem;margin-top:0.5rem">
{{if eq .Status "运行中"}}
<form action="/stop/{{.Name}}" method="post"><button class="btn btn-warning" style="width:100%"><i class="fa-solid fa-stop"></i>停止</button></form>
<form action="/restart/{{.Name}}" method="post"><button class="btn btn-primary" style="width:100%"><i class="fa-solid fa-rotate"></i>重启</button></form>
{{else}}
<form action="/start/{{.Name}}" method="post"><button class="btn btn-success" style="width:100%"><i class="fa-solid fa-play"></i>启动</button></form>
<a href="/" class="btn btn-ghost" style="width:100%"><i class="fa-solid fa-arrow-left"></i>返回</a>
{{end}}
</div>
</div>
<div style="font-size:0.68rem;color:var(--text2);margin-top:0.8rem;text-align:center"><i class="fa-solid fa-clock" style="margin-right:0.2rem"></i>创建时间：{{.Created}}{{if .Version}} | 版本：{{.Version}}{{end}}{{if .SourceURL}}<br><i class="fa-solid fa-link" style="margin-right:0.2rem"></i>来源：<a href="{{.SourceURL}}" target="_blank" style="color:#60a5fa;text-decoration:none">{{.SourceURL}}</a>{{end}}</div>
</div>
</div>
</div>
` + layoutJS + `
</body>
</html>`

var htmlSource = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width,initial-scale=1.0">
<title>源管理 - LightPanel</title>
<style>
` + layoutCSS + `
</style>
</head>
<body>
<div class="bg-layer"{{if getBgUrl}} style="background-image:url('{{getBgUrl}}')"{{end}}></div>
` + sidebarHTML("/source") + `
<div class="main">
` + topbarHTML("源管理") + `
<div class="content">
<div class="glass" style="padding:1rem;margin-bottom:0.8rem">
<h3 style="font-size:0.9rem;font-weight:600;color:var(--text);margin-bottom:0.8rem"><i class="fa-solid fa-plus" style="color:#10b981;margin-right:0.3rem"></i>添加源</h3>
<form action="/source/add" method="post" style="display:grid;grid-template-columns:1fr 2fr auto;gap:0.5rem;align-items:end">
<input name="name" placeholder="源名称" required class="input">
<input name="url" placeholder="JSON 地址" required class="input">
<button class="btn btn-primary" style="height:38px"><i class="fa-solid fa-plus"></i>添加</button>
</form>
</div>
<div style="display:grid;gap:0.5rem">
{{range $i,$s := .}}
<div class="card" style="display:flex;justify-content:space-between;align-items:center">
<div><span style="font-size:0.85rem;font-weight:600;color:var(--text)">{{$s.Name}}</span><br><span style="font-size:0.68rem;color:var(--text2);word-break:break-all">{{$s.URL}}</span></div>
<form action="/source/del/{{$i}}" method="post"><button class="btn btn-danger btn-sm"><i class="fa-solid fa-trash"></i></button></form>
</div>
{{end}}
</div>
</div>
</div>
</div>
` + layoutJS + `
</body>
</html>`

var htmlSetting = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width,initial-scale=1.0">
<title>设置 - LightPanel</title>
<style>
` + layoutCSS + `
</style>
</head>
<body>
<div class="bg-layer"{{if getBgUrl}} style="background-image:url('{{getBgUrl}}')"{{end}}></div>
` + sidebarHTML("/setting") + `
<div class="main">
` + topbarHTML("设置") + `
<div class="content">
{{if eq .Err "password"}}<div class="alert alert-error"><i class="fa-solid fa-triangle-exclamation" style="margin-right:0.3rem"></i>原密码错误</div>{{end}}
{{if eq .Err "password_weak"}}<div class="alert alert-error"><i class="fa-solid fa-triangle-exclamation" style="margin-right:0.3rem"></i>密码长度至少6位</div>{{end}}
{{if eq .Err "username_short"}}<div class="alert alert-error"><i class="fa-solid fa-triangle-exclamation" style="margin-right:0.3rem"></i>用户名长度至少3位</div>{{end}}
{{if eq .Err "missing_pwd"}}<div class="alert alert-error"><i class="fa-solid fa-triangle-exclamation" style="margin-right:0.3rem"></i>请输入原密码以验证身份</div>{{end}}
{{if eq .Msg "1"}}<div class="alert alert-success"><i class="fa-solid fa-check" style="margin-right:0.3rem"></i>设置已保存</div>{{end}}
<div class="glass" style="padding:1rem;margin-bottom:0.8rem">
<h3 style="font-size:0.9rem;font-weight:600;color:var(--text);margin-bottom:0.8rem"><i class="fa-solid fa-user" style="color:#f59e0b;margin-right:0.3rem"></i>账号设置</h3>
<form action="/setting/save-account" method="post" style="display:grid;gap:0.5rem">
<div><label style="font-size:0.7rem;color:var(--text2);margin-bottom:0.2rem;display:block">当前用户名</label><input name="current_username" value="{{.Username}}" readonly class="input" style="opacity:0.6"></div>
<div><label style="font-size:0.7rem;color:var(--text2);margin-bottom:0.2rem;display:block">新用户名（留空不修改）</label><input name="new_username" placeholder="至少3位" class="input"></div>
<div><label style="font-size:0.7rem;color:var(--text2);margin-bottom:0.2rem;display:block">原密码（验证身份）</label><input name="old" type="password" class="input"></div>
<div><label style="font-size:0.7rem;color:var(--text2);margin-bottom:0.2rem;display:block">新密码（留空不修改）</label><input name="new" type="password" placeholder="至少6位" class="input"></div>
<button class="btn btn-primary" style="width:100%"><i class="fa-solid fa-save"></i>保存账号</button>
</form>
</div>
<div class="glass" style="padding:1rem">
<h3 style="font-size:0.9rem;font-weight:600;color:var(--text);margin-bottom:0.8rem"><i class="fa-solid fa-image" style="color:#f59e0b;margin-right:0.3rem"></i>背景设置</h3>
<form action="/setting/save-bg" method="post" style="display:grid;gap:0.5rem">
<div><label style="font-size:0.7rem;color:var(--text2);margin-bottom:0.2rem;display:block">背景类型</label>
<select name="bg_type" class="input">
<option value="gradient"{{if eq .BgType "gradient"}} selected{{end}}>渐变</option>
<option value="image"{{if eq .BgType "image"}} selected{{end}}>图片</option>
</select></div>
<div><label style="font-size:0.7rem;color:var(--text2);margin-bottom:0.2rem;display:block">图片地址</label><input name="bg_url" value="{{.BgUrl}}" placeholder="图片 URL" class="input"></div>
<button class="btn btn-primary" style="width:100%"><i class="fa-solid fa-save"></i>保存背景</button>
</form>
</div>
<div class="glass" style="padding:1rem;margin-top:0.8rem">
<h3 style="font-size:0.9rem;font-weight:600;color:var(--text);margin-bottom:0.8rem"><i class="fa-solid fa-icons" style="color:#f59e0b;margin-right:0.3rem"></i>侧栏图标</h3>
<form action="/setting/save-logo" method="post" style="display:grid;gap:0.5rem">
<div><label style="font-size:0.7rem;color:var(--text2);margin-bottom:0.2rem;display:block">Logo 地址（留空使用默认图标）</label><input name="logo_url" value="{{.LogoUrl}}" placeholder="图片 URL" class="input"></div>
<button class="btn btn-primary" style="width:100%"><i class="fa-solid fa-save"></i>保存图标</button>
</form>
</div>
<div class="glass" style="padding:1rem;margin-top:0.8rem">
<h3 style="font-size:0.9rem;font-weight:600;color:var(--text);margin-bottom:0.8rem"><i class="fa-solid fa-database" style="color:#f59e0b;margin-right:0.3rem"></i>备份与恢复</h3>
<div style="display:grid;grid-template-columns:1fr 1fr;gap:0.5rem">
<form action="/setting/backup" method="post">
<button class="btn btn-primary" style="width:100%"><i class="fa-solid fa-download"></i>备份数据</button>
</form>
<form action="/setting/restore" method="post" enctype="multipart/form-data" onsubmit="return confirm('恢复将覆盖当前数据，确定继续？')">
<div style="display:flex;gap:0.3rem;align-items:center">
<input type="file" name="backup_file" accept=".tar.gz" class="input" style="padding:0.4rem;font-size:0.72rem">
<button class="btn btn-warning" style="white-space:nowrap"><i class="fa-solid fa-upload"></i>恢复</button>
</div>
</form>
</div>
</div>
<div class="glass" style="padding:1rem;margin-top:0.8rem">
<h3 style="font-size:0.9rem;font-weight:600;color:var(--text);margin-bottom:0.8rem"><i class="fa-solid fa-bell" style="color:#f59e0b;margin-right:0.3rem"></i>Bark 通知</h3>
<form action="/setting/save-bark" method="post" style="display:grid;gap:0.5rem">
<label class="check-label"><input type="checkbox" name="bark_enabled"{{if .BarkEnabled}} checked{{end}}><div class="check-box"><i class="fa-solid fa-check"></i></div><span style="color:var(--text2);font-size:0.78rem">启用 Bark 推送（应用崩溃/停止时通知）</span></label>
<div><label style="font-size:0.7rem;color:var(--text2);margin-bottom:0.2rem;display:block">设备 Key（或完整 URL）</label><input name="bark_device" value="{{.BarkDevice}}" placeholder="如 xxxxxxx 或 https://api.day.app/xxxxxxx" class="input"></div>
<div><label style="font-size:0.7rem;color:var(--text2);margin-bottom:0.2rem;display:block">分组（可选）</label><input name="bark_group" value="{{.BarkGroup}}" placeholder="LightPanel" class="input"></div>
<button class="btn btn-primary" style="width:100%"><i class="fa-solid fa-save"></i>保存通知</button>
</form>
</div>
</div>
</div>
</div>
` + layoutJS + `
</body>
</html>`

var htmlSystem = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width,initial-scale=1.0">
<title>系统监控 - LightPanel</title>
<style>
` + layoutCSS + `
</style>
</head>
<body>
<div class="bg-layer"{{if getBgUrl}} style="background-image:url('{{getBgUrl}}')"{{end}}></div>
` + sidebarHTML("/system") + `
<div class="main">
` + topbarHTML("系统监控") + `
<div class="content">
<div style="display:grid;grid-template-columns:repeat(4,1fr);gap:0.6rem;margin-bottom:1rem">
<div class="card stat-card"><div class="stat-value">{{.Cpu}}%</div><div class="stat-label"><i class="fa-solid fa-microchip" style="margin-right:0.2rem"></i>CPU</div></div>
<div class="card stat-card"><div class="stat-value">{{.Mem}}%</div><div class="stat-label"><i class="fa-solid fa-memory" style="margin-right:0.2rem"></i>内存</div></div>
<div class="card stat-card"><div class="stat-value">{{.Disk}}%</div><div class="stat-label"><i class="fa-solid fa-hard-drive" style="margin-right:0.2rem"></i>磁盘</div></div>
<div class="card stat-card"><div class="stat-value">{{.ProcNum}}</div><div class="stat-label"><i class="fa-solid fa-list" style="margin-right:0.2rem"></i>进程</div></div>
</div>
<div class="glass" style="padding:1rem">
<div style="display:flex;justify-content:space-between;align-items:center;margin-bottom:0.8rem">
<h3 style="font-size:0.9rem;font-weight:600;color:var(--text)">进程列表</h3>
<span style="font-size:0.7rem;color:var(--text2)">自动刷新 5s</span>
</div>
<div class="proc-wrap">
<table class="proc-table">
<thead><tr><th>名称</th><th>PID</th><th>CPU</th><th>内存</th><th>操作</th></tr></thead>
<tbody>
{{range .Procs}}
<tr>
<td>{{.Name}}</td>
<td>{{.PID}}</td>
<td>{{.Cpu}}%</td>
<td>{{.Mem}}%</td>
<td><form action="/kill/{{.PID}}" method="post"><button class="btn" style="padding:0.3rem 0.8rem;font-size:0.7rem;background:#ef4444;color:#fff;border:none;border-radius:7px;cursor:pointer">关闭</button></form></td>
</tr>
{{end}}
</tbody>
</table>
</div>
</div>
</div>
</div>
</div>
` + layoutJS + `
<script>setInterval(function(){location.reload()},5000)</script>
</body>
</html>`

var htmlLog = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width,initial-scale=1.0">
<title>运行日志 - LightPanel</title>
<style>
` + layoutCSS + `
.log-box{background:rgba(0,0,0,0.3);border-radius:10px;padding:0.8rem;font-family:'JetBrains Mono',monospace;font-size:0.72rem;color:#e2e8f0;max-height:70vh;overflow-y:auto;white-space:pre-wrap;word-break:break-all;line-height:1.6}
</style>
</head>
<body>
<div class="bg-layer"{{if getBgUrl}} style="background-image:url('{{getBgUrl}}')"{{end}}></div>
` + sidebarHTML("/log/") + `
<div class="main">
` + topbarHTML("运行日志") + `
<div class="content">
<div style="display:flex;gap:0.5rem;margin-bottom:0.8rem">
<a href="/" class="btn btn-ghost"><i class="fa-solid fa-arrow-left"></i>返回</a>
<form action="/log/clear/{{.Name}}" method="post"><button class="btn btn-danger"><i class="fa-solid fa-trash"></i>清空日志</button></form>
<span style="margin-left:auto;font-size:0.72rem;color:var(--text2)" id="logCount">{{.LineCount}} 行</span>
</div>
<div class="log-box" id="logBox">{{escape .Log}}</div>
</div>
</div>
</div>
` + layoutJS + `
</body>
</html>`

var htmlScriptAnalyze = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width,initial-scale=1.0">
<title>脚本分析 - LightPanel</title>
<style>
` + layoutCSS + `
.tag{display:inline-block;padding:0.2rem 0.5rem;border-radius:5px;font-size:0.7rem;font-family:monospace;margin:0.15rem}
.tag-dep{background:rgba(245,158,11,0.1);color:#fbbf24;border:1px solid rgba(245,158,11,0.2)}
.tag-port{background:rgba(16,185,129,0.1);color:#34d399;border:1px solid rgba(16,185,129,0.2)}
.tag-env{background:rgba(99,102,241,0.1);color:#a5b4fc;border:1px solid rgba(99,102,241,0.2)}
</style>
</head>
<body>
<div class="bg-layer"{{if getBgUrl}} style="background-image:url('{{getBgUrl}}')"{{end}}></div>
` + sidebarHTML("/analyze") + `
<div class="main">
` + topbarHTML("脚本分析") + `
<div class="content">
<div class="glass" style="padding:1rem">
<div style="display:flex;gap:0.5rem;margin-bottom:0.8rem">
<input id="scriptUrl" placeholder="输入脚本 URL" class="input" style="flex:1">
<button class="btn btn-primary" onclick="analyze()"><i class="fa-solid fa-magnifying-glass"></i>分析</button>
</div>
<div id="result" style="display:none">
<div style="margin-bottom:0.6rem"><h4 style="font-size:0.78rem;color:var(--text2);margin-bottom:0.3rem"><i class="fa-solid fa-cube" style="margin-right:0.2rem"></i>依赖</h4><div id="deps"></div></div>
<div style="margin-bottom:0.6rem"><h4 style="font-size:0.78rem;color:var(--text2);margin-bottom:0.3rem"><i class="fa-solid fa-network-wired" style="margin-right:0.2rem"></i>端口</h4><div id="ports"></div></div>
<div><h4 style="font-size:0.78rem;color:var(--text2);margin-bottom:0.3rem"><i class="fa-solid fa-gear" style="margin-right:0.2rem"></i>环境变量</h4><div id="envs"></div></div>
</div>
<div id="loading" style="display:none;text-align:center;padding:1.5rem;color:var(--text2)"><i class="fa-solid fa-spinner fa-spin" style="font-size:1.2rem;margin-bottom:0.3rem"></i><p style="font-size:0.78rem">正在分析...</p></div>
</div>
</div>
</div>
</div>
` + layoutJS + `
<script>
function analyze(){
var url=document.getElementById('scriptUrl').value.trim();
if(!url)return;
document.getElementById('result').style.display='none';
document.getElementById('loading').style.display='block';
fetch('/api/analyze?url='+encodeURIComponent(url)).then(function(r){return r.json()}).then(function(data){
document.getElementById('loading').style.display='none';
document.getElementById('result').style.display='block';
document.getElementById('deps').innerHTML=data.deps&&data.deps.length?data.deps.map(function(d){return '<span class="tag tag-dep">'+d+'</span>'}).join(''):'<span style="color:var(--text2);font-size:0.72rem">未检测到</span>';
document.getElementById('ports').innerHTML=data.ports&&data.ports.length?data.ports.map(function(p){return '<span class="tag tag-port">'+p+'</span>'}).join(''):'<span style="color:var(--text2);font-size:0.72rem">未检测到</span>';
document.getElementById('envs').innerHTML=data.env&&data.env.length?data.env.map(function(e){return '<span class="tag tag-env">'+e+'</span>'}).join(''):'<span style="color:var(--text2);font-size:0.72rem">未检测到</span>';
}).catch(function(){document.getElementById('loading').style.display='none';alert('分析失败')});
}
</script>
</body>
</html>`
