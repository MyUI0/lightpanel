package handlers

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
{{.Sidebar}}
<div class="main">
{{.Topbar}}
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
<input type="hidden" name="csrf_token" value="{{$.CSRFToken}}">
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
{{end}}
<a href="/" class="btn btn-ghost" style="width:100%"><i class="fa-solid fa-arrow-left"></i>返回</a>
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
.edit-form{display:none;grid-template-columns:1fr 2fr auto auto;gap:0.8rem;align-items:end;margin-top:0.8rem;padding:0.6rem;background:var(--glass);border-radius:10px}
.edit-form.active{display:grid}
</style>
</head>
<body>
<div class="bg-layer"{{if getBgUrl}} style="background-image:url('{{getBgUrl}}')"{{end}}></div>
{{.Sidebar}}
<div class="main">
{{.Topbar}}
<div class="content">
{{if eq .Message "test_ok"}}<div class="alert alert-success"><i class="fa-solid fa-check" style="margin-right:0.3rem"></i>源连接成功</div>{{end}}
{{if eq .Error "test_failed"}}<div class="alert alert-error"><i class="fa-solid fa-triangle-exclamation" style="margin-right:0.3rem"></i>源连接失败</div>{{end}}
{{if eq .Error "test_empty"}}<div class="alert alert-error"><i class="fa-solid fa-triangle-exclamation" style="margin-right:0.3rem"></i>源返回数据为空</div>{{end}}
<div style="margin-bottom:0.8rem"><a href="/store" class="btn btn-ghost btn-sm"><i class="fa-solid fa-arrow-left"></i> 返回商店</a></div>
<div class="glass" style="padding:1rem;margin-bottom:0.8rem">
<h3 style="font-size:0.9rem;font-weight:600;color:var(--text);margin-bottom:0.8rem"><i class="fa-solid fa-plus" style="color:#10b981;margin-right:0.3rem"></i>添加源</h3>
<form action="/source/add" method="post" style="display:grid;grid-template-columns:1fr 2fr auto;gap:0.5rem;align-items:end">
<input name="name" placeholder="源名称" required class="input">
<input name="url" placeholder="JSON 地址" required class="input">
<button class="btn btn-primary" style="height:38px"><i class="fa-solid fa-plus"></i>添加</button>
</form>
</div>
<div style="display:grid;gap:0.5rem">
{{range $i,$s := .Sources}}
<div class="card">
<div style="display:flex;justify-content:space-between;align-items:center">
<div><span style="font-size:0.85rem;font-weight:600;color:var(--text)">{{$s.Name}}</span><br><span style="font-size:0.68rem;color:var(--text2);word-break:break-all">{{$s.URL}}</span></div>
<div style="display:flex;gap:0.3rem">
<a href="/source/test/{{$i}}" class="btn btn-ghost btn-sm"><i class="fa-solid fa-plug"></i> 测试</a>
<button class="btn btn-ghost btn-sm" onclick="toggleEditForm({{$i}})"><i class="fa-solid fa-pen"></i> 编辑</button>
<form action="/source/del/{{$i}}" method="post"><button class="btn btn-danger btn-sm"><i class="fa-solid fa-trash"></i> 删除</button></form>
</div>
</div>
<form id="edit-{{$i}}" class="edit-form" action="/source/edit/{{$i}}" method="post">
<input name="name" value="{{$s.Name}}" required class="input">
<input name="url" value="{{$s.URL}}" required class="input">
<button class="btn btn-primary btn-sm"><i class="fa-solid fa-check"></i> 保存</button>
<button type="button" class="btn btn-ghost btn-sm" onclick="toggleEditForm({{$i}})"><i class="fa-solid fa-xmark"></i> 取消</button>
</form>
</div>
{{end}}
</div>
</div>
</div>
</div>
` + layoutJS + `
</body>
</html>`
