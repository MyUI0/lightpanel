package handlers

var htmlIndex = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width,initial-scale=1.0">
<title>仪表盘 - 朱雀面板</title>
<style>
` + layoutCSS + `
</style>
</head>
<body>
<div class="bg-layer"{{if .BgUrl}} style="background-image:url('{{.BgUrl}}')"{{end}}></div>
{{.Sidebar}}
<div class="main">
{{.Topbar}}
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
<div style="display:grid;grid-template-columns:repeat(4,1fr);gap:0.6rem;margin-bottom:1rem">
<div class="card stat-card"><div class="stat-value">{{.Cpu}}%</div><div class="stat-label"><i class="fa-solid fa-microchip" style="margin-right:0.2rem"></i>CPU</div></div>
<div class="card stat-card"><div class="stat-value">{{.Mem}}%</div><div class="stat-label"><i class="fa-solid fa-memory" style="margin-right:0.2rem"></i>内存</div></div>
<div class="card stat-card"><div class="stat-value">{{.Disk}}%</div><div class="stat-label"><i class="fa-solid fa-hard-drive" style="margin-right:0.2rem"></i>磁盘</div></div>
<div class="card stat-card"><div class="stat-value">{{.ProcNum}}</div><div class="stat-label"><i class="fa-solid fa-list" style="margin-right:0.2rem"></i>进程</div></div>
</div>
<div style="font-size:0.7rem;color:var(--text2);margin-bottom:1rem;text-align:center"><i class="fa-solid fa-clock" style="margin-right:0.2rem"></i>运行时间: {{.Uptime}}</div>
</div>
</div>
</div>
` + layoutJS + `
</body>
</html>`
