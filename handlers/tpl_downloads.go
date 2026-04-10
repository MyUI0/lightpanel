package handlers

var htmlDownloads = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width,initial-scale=1.0">
<title>下载管理 - 朱雀面板</title>
<meta name="csrf" content="{{.CSRFToken}}">
<style>
` + layoutCSS + `
.badge.downloading{background:rgba(59,130,246,0.1);color:#60a5fa;border:1px solid rgba(59,130,246,0.2)}
.badge.paused{background:rgba(245,158,11,0.1);color:#fbbf24;border:1px solid rgba(245,158,11,0.2)}
.badge.completed{background:rgba(16,185,129,0.1);color:#34d399;border:1px solid rgba(16,185,129,0.2)}
.badge.error{background:rgba(239,68,68,0.1);color:#f87171;border:1px solid rgba(239,68,68,0.2)}
</style>
</head>
<body>
<div class="bg-layer"{{if getBgUrl}} style="background-image:url('{{getBgUrl}}')"{{end}}></div>
{{.Sidebar}}
<div class="main">
{{.Topbar}}
<div class="content">
{{if .Tasks}}
<div style="display:flex;justify-content:space-between;align-items:center;margin-bottom:0.8rem">
<h3 style="font-size:0.9rem;font-weight:600;color:var(--text)">当前任务</h3>
</div>
<div id="taskList">
{{range .Tasks}}
<div class="card" id="task-{{.ID}}" style="margin-bottom:0.5rem">
<div style="display:flex;justify-content:space-between;align-items:center;margin-bottom:0.3rem">
<div><span style="font-size:0.88rem;font-weight:600;color:var(--text)">{{.Name}}</span><span class="badge badge-{{.Status}}" style="margin-left:0.4rem">{{if eq .Status "downloading"}}下载中{{end}}{{if eq .Status "paused"}}已暂停{{end}}{{if eq .Status "deploying"}}部署中...{{end}}{{if eq .Status "completed"}}已完成{{end}}{{if eq .Status "error"}}失败{{end}}</span></div>
<div style="display:flex;gap:0.3rem">
{{if eq .Status "downloading"}}<button class="btn btn-warning btn-sm" onclick="dlAction('{{.ID}}','pause')"><i class="fa-solid fa-pause"></i> 暂停</button>{{end}}
{{if eq .Status "paused"}}<button class="btn btn-success btn-sm" onclick="dlAction('{{.ID}}','resume')"><i class="fa-solid fa-play"></i> 继续</button>{{end}}
{{if eq .Status "completed"}}<button class="btn btn-primary btn-sm" onclick="dlAction('{{.ID}}','install')"><i class="fa-solid fa-rocket"></i> 部署</button>{{end}}
<button class="btn btn-danger btn-sm" onclick="if(confirm('确定删除此下载任务？'))dlAction('{{.ID}}','delete')"><i class="fa-solid fa-trash"></i> 删除</button>
</div>
</div>
<div style="font-size:0.68rem;color:var(--text2);word-break:break-all;margin-bottom:0.3rem">{{.URL}}</div>
<div class="progress-bar"><div class="progress-fill" style="width:{{.Progress}}%"></div></div>
<div style="font-size:0.68rem;color:var(--text2);display:flex;justify-content:space-between"><span class="progress-pct">{{.Progress}}%</span><span class="progress-size">{{formatSize .Downloaded}} / {{formatSize .Size}}</span></div>
</div>
{{end}}
{{if eq (len .Tasks) 0}}
<div class="glass" style="padding:2rem;text-align:center;margin-bottom:1rem"><i class="fa-solid fa-inbox" style="font-size:2rem;color:rgba(255,255,255,0.1);margin-bottom:0.5rem"></i><p style="color:var(--text2);font-size:0.8rem">暂无下载任务</p></div>
{{end}}
{{end}}
{{if .History}}
<div style="display:flex;justify-content:space-between;align-items:center;margin-bottom:0.8rem;margin-top:1rem">
<h3 style="font-size:0.9rem;font-weight:600;color:var(--text)">下载记录</h3>
</div>
<div style="display:grid;gap:0.4rem">
{{range .History}}
<div class="card" style="padding:0.6rem 0.8rem">
<div style="display:flex;justify-content:space-between;align-items:center">
<span style="font-size:0.8rem;font-weight:600;color:var(--text)">{{.Name}}{{if .Version}} <span style="font-size:0.65rem;color:var(--text2)">v{{.Version}}</span>{{end}}</span>
{{if .Installed}}<span class="badge badge-running">已部署</span>{{else}}<span class="badge badge-stopped">未部署</span>{{end}}
</div>
<div style="font-size:0.65rem;color:var(--text2);margin-top:0.2rem">{{formatTime .Timestamp}}</div></div>
{{end}}
{{if eq (len .History) 0}}
<div class="glass" style="padding:1.5rem;text-align:center"><i class="fa-solid fa-clock-rotate-left" style="font-size:1.5rem;color:rgba(255,255,255,0.1);margin-bottom:0.5rem"></i><p style="color:var(--text2);font-size:0.8rem">暂无下载记录</p></div>
{{end}}
</div>
{{end}}
</div>
</div>
</div>
` + layoutJS + `
function formatSize(b){if(b<=0)return '0 B';var u=['B','KB','MB','GB','TB'];var i=0;while(b>=1024&&i<u.length-1){b/=1024;i++}return b.toFixed(1)+' '+u[i]}
function dlAction(id,act){var csrf=document.querySelector('meta[name="csrf"]');fetch('/dl/action/'+id+'/'+act,{method:'POST',headers:csrf?{'X-CSRF-Token':csrf.content}:{}}).then(function(){location.reload()})}
setInterval(function(){fetch('/dl/api').then(function(r){return r.json()}).then(function(tasks){if(tasks&&tasks.forEach){tasks.forEach(function(t){var card=document.getElementById('task-'+t.id);if(!card)return;var fill=card.querySelector('.progress-fill');if(fill)fill.style.width=t.progress+'%';var pct=card.querySelector('.progress-pct');if(pct)pct.textContent=t.progress+'%';var sz=card.querySelector('.progress-size');if(sz&&t.size>0)sz.textContent=formatSize(t.downloaded)+' / '+formatSize(t.size)})}})},1000);
</script>
</body>
</html>`