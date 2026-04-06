package handlers

var htmlStore = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width,initial-scale=1.0">
<title>应用商店 - LightPanel</title>
<style>
` + layoutCSS + `
.store-item{display:flex;flex-direction:column;align-items:center;text-align:center;padding:1rem;gap:0.5rem}
.store-icon{width:44px;height:44px;border-radius:10px;object-fit:cover;border:1px solid var(--border)}
.source-tabs{display:flex;gap:0.4rem;margin-bottom:1rem;flex-wrap:wrap}
.source-tab{padding:0.4rem 0.8rem;border-radius:8px;font-size:0.75rem;text-decoration:none;transition:all 0.15s}
.source-tab.active{background:linear-gradient(135deg,#6366f1,#7c3aed);color:#fff}
.source-tab:not(.active){background:var(--card);color:var(--text2);border:1px solid var(--border)}
</style>
</head>
<body>
<div class="bg-layer"{{if getBgUrl}} style="background-image:url('{{getBgUrl}}')"{{end}}></div>
{{.Sidebar}}
<div class="main">
{{.Topbar}}
<div class="content">
<div class="source-tabs">
{{range $i,$s := .Sources}}
{{if eq $i $.Active}}<span class="source-tab active">{{$s.Name}}</span>{{else}}<a href="/store?source={{$i}}" class="source-tab">{{$s.Name}}</a>{{end}}
{{end}}
{{if .Sources}}<a href="/source" class="source-tab" style="border-style:dashed"><i class="fa-solid fa-pen"></i> 编辑源</a>{{end}}
{{if .StoreErr}}<a href="/source" class="source-tab" style="color:#f87171"><i class="fa-solid fa-triangle-exclamation"></i> 源配置</a>{{end}}
</div>
{{if .StoreErr}}<div class="alert alert-{{if eq .StoreErrType "network"}}error{{else if eq .StoreErrType "http"}}warning{{else}}error{{end}}" style="margin-bottom:1rem"><i class="fa-solid fa-circle-exclamation"></i> {{.StoreErr}}</div>{{end}}
{{if .DeployErr}}<div class="alert alert-error" style="margin-bottom:1rem"><i class="fa-solid fa-circle-exclamation"></i> 部署失败: {{.DeployErr}}</div>{{end}}
<div style="margin-bottom:1rem">
<input type="text" id="storeSearch" placeholder="搜索应用..." class="input" style="width:100%" oninput="filterStore()">
</div>
<div style="display:grid;grid-template-columns:repeat(3,1fr);gap:0.5rem" id="storeList">
{{range $i,$a := .Apps}}
<div class="card store-item" data-name="{{tolower $a.Name}}" data-desc="{{tolower $a.Desc}}">
<img src="{{$a.Icon}}" class="store-icon" referrerpolicy="no-referrer" onerror="this.style.display='none'" style="width:48px;height:48px">
<div style="flex:1;min-width:0;width:100%">
<h3 style="font-size:0.85rem;font-weight:600;color:var(--text)">{{$a.Name}}{{if $a.Version}}<span style="font-size:0.6rem;background:var(--card);color:var(--text2);padding:0.1rem 0.3rem;margin-left:0.3rem;border-radius:4px">v{{$a.Version}}</span>{{end}}</h3>
<p style="font-size:0.68rem;color:var(--text2);margin-top:0.3rem;display:-webkit-box;-webkit-line-clamp:2;-webkit-box-orient:vertical;overflow:hidden">{{$a.Desc}}</p>
<p style="font-size:0.6rem;color:var(--text2);margin-top:0.3rem"><i class="fa-solid fa-user" style="margin-right:0.2rem"></i>{{$a.Author}}</p>
</div>
{{if index $.Deployed $a.Name}}
<button class="btn btn-ghost" style="width:100%;opacity:0.6" disabled><i class="fa-solid fa-check"></i> 已部署</button>
{{else if $.StoreErr}}
<button class="btn btn-ghost" style="width:100%;opacity:0.5" disabled><i class="fa-solid fa-triangle-exclamation"></i> 获取失败</button>
{{else}}
<form action="/install/{{$i}}?source={{$.Active}}" method="post" class="install-form" style="width:100%">
<input type="hidden" name="csrf_token" value="{{$.CSRFToken}}">
<button class="btn btn-success" style="width:100%"><i class="fa-solid fa-download"></i>部署</button>
</form>
{{end}}
</div>
{{end}}
</div>
</div>
</div>
</div>
` + layoutJS + `
<script>
function filterStore(){
var kw=document.getElementById('storeSearch').value.toLowerCase();
document.querySelectorAll('.store-item').forEach(function(el){var name=el.getAttribute('data-name')||'';var desc=el.getAttribute('data-desc')||'';el.style.display=(kw&&!name.includes(kw)&&!desc.includes(kw))?'none':''})}
document.querySelectorAll('.install-form').forEach(function(f){f.addEventListener('submit',function(e){var b=f.querySelector('button');if(b&&b.disabled){e.preventDefault();return}})});
</script>
</body>
</html>`

var htmlStoreParams = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width,initial-scale=1.0">
<title>设置参数 - LightPanel</title>
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
<p style="color:var(--text);font-size:0.85rem;margin-bottom:0.8rem">应用 <strong>{{.App.Name}}</strong> 需要设置运行参数</p>
<div style="font-size:0.72rem;color:var(--text2);margin-bottom:1rem;padding:0.5rem;background:var(--card);border-radius:7px;word-break:break-all">
检测到下载链接: {{.TestedURL}}
</div>
<form action="/install/confirm/{{.Index}}?source={{.SrcIdx}}" method="post">
<div><label style="font-size:0.7rem;color:var(--text2);margin-bottom:0.2rem;display:block">{{if .App.ParamsHint}}{{.App.ParamsHint}}{{else}}运行参数{{end}}</label>
<input name="user_params" placeholder="如: -e https://xxx -t token" class="input" style="width:100%"></div>
<div style="display:grid;grid-template-columns:1fr 1fr;gap:0.5rem;margin-top:1rem">
<button type="submit" class="btn btn-primary" style="width:100%"><i class="fa-solid fa-download"></i> 确认部署</button>
<a href="/store" class="btn btn-ghost" style="width:100%;text-align:center"><i class="fa-solid fa-xmark"></i> 取消</a>
</div>
</form>
</div>
</div>
</div>
</div>
` + layoutJS + `
</body>
</html>`
