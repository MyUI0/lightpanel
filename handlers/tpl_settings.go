package handlers

var htmlSetting = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width,initial-scale=1.0">
<title>设置 - 朱雀面板</title>
<style>
` + layoutCSS + `
</style>
</head>
<body>
<div class="bg-layer"{{if getBgUrl}} style="background-image:url('{{getBgUrl}}')"{{end}}></div>
{{.Sidebar}}
<div class="main">
{{.Topbar}}
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
<label class="file-input-btn">
<input type="file" name="backup_file" accept=".tar.gz,.tar" class="file-input" style="display:none" onchange="var f=this.files[0];this.closest('form').querySelector('.file-name').textContent=f?f.name:''">
<span class="file-name btn btn-warning" style="white-space:nowrap"><i class="fa-solid fa-upload"></i>选择文件</span>
</label>
<button type="submit" class="btn btn-warning" style="white-space:nowrap"><i class="fa-solid fa-check"></i>恢复</button>
</div>
</form>
</div>
</div>
<div class="glass" style="padding:1rem;margin-top:0.8rem">
<h3 style="font-size:0.9rem;font-weight:600;color:var(--text);margin-bottom:0.8rem"><i class="fa-solid fa-bell" style="color:#f59e0b;margin-right:0.3rem"></i>Bark 通知</h3>
<form action="/setting/save-bark" method="post" style="display:grid;gap:0.5rem">
<label class="check-label"><input type="checkbox" name="bark_enabled"{{if .BarkEnabled}} checked{{end}}><div class="check-box"><i class="fa-solid fa-check"></i></div><span style="color:var(--text2);font-size:0.78rem">启用 Bark 推送（应用崩溃/停止时通知）</span></label>
<div><label style="font-size:0.7rem;color:var(--text2);margin-bottom:0.2rem;display:block">设备 Key（或完整 URL）</label><input name="bark_device" value="{{.BarkDevice}}" placeholder="如 xxxxxxx 或 https://api.day.app/xxxxxxx" class="input"></div>
<div><label style="font-size:0.7rem;color:var(--text2);margin-bottom:0.2rem;display:block">分组（可选）</label><input name="bark_group" value="{{.BarkGroup}}" placeholder="朱雀面板" class="input"></div>
<button class="btn btn-primary" style="width:100%"><i class="fa-solid fa-save"></i>保存通知</button>
</form>
</div>
<div class="glass" style="padding:1rem;margin-top:0.8rem">
<h3 style="font-size:0.9rem;font-weight:600;color:var(--text);margin-bottom:0.8rem"><i class="fa-solid fa-file-circle-plus" style="color:#f59e0b;margin-right:0.3rem"></i>自定义页面</h3>
<form action="/setting/upload-page" method="post" enctype="multipart/form-data" style="display:grid;gap:0.5rem">
<div style="display:flex;gap:0.3rem;align-items:center">
<label class="file-input-btn">
<input type="file" name="page_file" accept=".html" class="file-input" style="display:none" onchange="var f=this.files[0];this.closest('form').querySelector('.file-name').textContent=f?f.name:''">
<span class="file-name btn btn-ghost" style="white-space:nowrap"><i class="fa-solid fa-upload"></i>选择HTML文件</span>
</label>
<input name="filename" placeholder="页面名称（选填）" class="input" style="flex:1">
<button type="submit" class="btn btn-primary"><i class="fa-solid fa-plus"></i>上传</button>
</div>
</form>
{{if .CustomPages}}
<div style="margin-top:0.8rem">
<p style="font-size:0.7rem;color:var(--text2);margin-bottom:0.5rem">已添加的页面：</p>
<div style="display:grid;gap:0.3rem">
{{range .CustomPages}}
<div style="display:flex;justify-content:space-between;align-items:center;padding:0.4rem 0.6rem;background:var(--card);border-radius:7px">
<span style="font-size:0.78rem;color:var(--text)">{{.}}</span>
<a href="/page/{{.}}" class="btn btn-ghost btn-sm"><i class="fa-solid fa-eye"></i> 查看</a>
<form action="/page/del/{{.}}" method="post" style="display:inline"><button class="btn btn-danger btn-sm"><i class="fa-solid fa-trash"></i></button></form>
</div>
{{end}}
</div>
</div>
{{end}}
</div>
</div>
</div>
</div>
` + layoutJS + `
</body>
</html>`

var htmlLogList = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width,initial-scale=1.0">
<title>运行日志 - 朱雀面板</title>
<style>
` + layoutCSS + `
</style>
</head>
<body>
<div class="bg-layer"{{if getBgUrl}} style="background-image:url('{{getBgUrl}}')"{{end}}></div>
{{.Sidebar}}
<div class="main">
{{.Topbar}}
<div class="content">
<div class="glass" style="padding:1rem">
<h3 style="font-size:0.9rem;font-weight:600;color:var(--text);margin-bottom:0.8rem"><i class="fa-solid fa-file-lines" style="margin-right:0.3rem"></i>选择应用查看日志</h3>
<div style="display:grid;gap:0.4rem">
{{range .Apps}}
<div style="display:flex;justify-content:space-between;align-items:center;padding:0.6rem 0.8rem;background:var(--card);border-radius:9px">
<span style="font-size:0.82rem;color:var(--text)">{{.}}</span>
<a href="/log/{{.}}" class="btn btn-ghost btn-sm"><i class="fa-solid fa-eye"></i> 查看日志</a>
</div>
{{end}}
</div>
{{if eq (len .Apps) 0}}
<div style="text-align:center;padding:2rem;color:var(--text2)">暂无应用</div>
{{end}}
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
<title>系统监控 - 朱雀面板</title>
<style>
` + layoutCSS + `
</style>
</head>
<body>
<div class="bg-layer"{{if getBgUrl}} style="background-image:url('{{getBgUrl}}')"{{end}}></div>
{{.Sidebar}}
<div class="main">
{{.Topbar}}
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
<td>{{printf "%.1f" .Cpu}}%</td>
<td>{{printf "%.2f" .Mem}}%</td>
<td><form action="/kill/{{.PID}}" method="post"><button class="btn" style="padding:0.3rem 0.8rem;font-size:0.7rem;background:#ef4444;color:#fff;border:none;border-radius:7px;cursor:pointer" onclick="return confirm('确定关闭进程 {{.Name}} (PID: {{.PID}}) 吗？')">关闭</button></form></td>
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
<script>setInterval(function(){if(!document.hidden)location.reload()},5000)</script>
</body>
</html>`

var htmlLog = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width,initial-scale=1.0">
<title>运行日志 - 朱雀面板</title>
<style>
` + layoutCSS + `
.log-box{background:rgba(15,15,35,0.6);border-radius:10px;padding:0.8rem;font-family:'JetBrains Mono',monospace;font-size:0.72rem;color:#e2e8f0;max-height:70vh;overflow-y:auto;white-space:pre-wrap;word-break:break-all;line-height:1.6;border:1px solid var(--border)}
</style>
</head>
<body>
<div class="bg-layer"{{if getBgUrl}} style="background-image:url('{{getBgUrl}}')"{{end}}></div>
{{.Sidebar}}
<div class="main">
{{.Topbar}}
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
<title>脚本分析 - 朱雀面板</title>
<style>
` + layoutCSS + `
.tag{display:inline-block;padding:0.2rem 0.5rem;border-radius:5px;font-size:0.7rem;font-family:monospace;margin:0.15rem}
.tag-dep{background:rgba(245,158,11,0.1);color:#fbbf24;border:1px solid rgba(245,158,11,0.2)}
.tag-port{background:rgba(16,185,129,0.1);color:#34d399;border:1px solid rgba(16,185,129,0.2)}
.tag-env{background:rgba(229,62,62,0.1);color:#fc8181;border:1px solid rgba(229,62,62,0.2)}
</style>
</head>
<body>
<div class="bg-layer"{{if getBgUrl}} style="background-image:url('{{getBgUrl}}')"{{end}}></div>
{{.Sidebar}}
<div class="main">
{{.Topbar}}
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
