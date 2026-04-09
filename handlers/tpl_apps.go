package handlers

var htmlApps = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width,initial-scale=1.0">
<title>管理应用 - 朱雀面板</title>
<style>
` + layoutCSS + `
.app-item{display:flex;align-items:center;gap:0.8rem}
.app-name{font-weight:600;color:var(--text);font-size:0.85rem}
.app-cmd{font-size:0.7rem;color:var(--text2);margin-top:0.15rem;max-width:300px;overflow:hidden;text-overflow:ellipsis;white-space:nowrap}
.app-grid{display:grid;grid-template-columns:1fr 1fr;gap:0.5rem}
.app-cb{margin-right:0.3rem}
.batch-bar{display:none;gap:0.3rem;align-items:center}
.batch-bar.active{display:flex}
</style>
</head>
<body>
<div class="bg-layer"{{if getBgUrl}} style="background-image:url('{{getBgUrl}}')"{{end}}></div>
{{.Sidebar}}
<div class="main">
{{.Topbar}}
<div class="content">
<input type="hidden" name="csrf_token" value="{{.CSRFToken}}">
<div style="display:flex;align-items:center;gap:0.5rem;margin-bottom:0.8rem;flex-wrap:wrap">
<div style="flex:1;min-width:200px">
<input type="text" id="appSearch" placeholder="搜索应用名称或命令..." class="input" style="padding:0.5rem 0.7rem;font-size:0.78rem">
</div>
<button class="btn btn-ghost btn-sm" id="selectAllBtn" onclick="toggleSelectMode()"><i class="fa-solid fa-check-double"></i> 全选</button>
<div class="batch-bar" id="batchBar">
<button class="btn btn-success btn-sm" onclick="batchAction('start')"><i class="fa-solid fa-play"></i> 启动</button>
<button class="btn btn-warning btn-sm" onclick="batchAction('stop')"><i class="fa-solid fa-stop"></i> 停止</button>
<button class="btn btn-danger btn-sm" onclick="batchAction('delete')"><i class="fa-solid fa-trash"></i> 删除</button>
<button class="btn btn-ghost btn-sm" onclick="toggleSelectMode()"><i class="fa-solid fa-xmark"></i> 取消</button>
<span id="selCount" style="font-size:0.7rem;color:var(--text2)">已选 0</span>
</div>
</div>
<div class="app-grid">
{{range $name, $app := .Apps}}
<div class="card app-item" data-name="{{tolower $name}}" data-cmd="{{tolower $app.Cmd}}">
<div style="flex:1;min-width:0">
<div style="display:flex;align-items:center;gap:0.5rem;flex-wrap:wrap;margin-bottom:0.5rem">
<span class="app-name app-cb" style="display:none"><input type="checkbox" class="app-cb" data-name="{{$name}}"></span>
<span>{{$name}}</span>
{{if eq $app.Status "运行中"}}<span class="badge badge-running"><span style="width:5px;height:5px;background:#34d399;border-radius:50%"></span>运行中</span>{{else}}<span class="badge badge-stopped"><span style="width:5px;height:5px;background:#f87171;border-radius:50%"></span>已停止</span>{{end}}
{{if $app.AutoStart}}<span class="badge" style="background:rgba(59,130,246,0.1);color:#60a5fa;border:1px solid rgba(59,130,246,0.2)">自启</span>{{end}}
{{if $app.Version}}<span class="badge" style="background:rgba(229,62,62,0.1);color:#fc8181;border:1px solid rgba(229,62,62,0.2)">v{{$app.Version}}</span>{{end}}
</div>
<div style="font-size:0.7rem;color:var(--text2);margin-bottom:0.5rem;word-break:break-all"><i class="fa-solid fa-terminal" style="margin-right:0.3rem"></i>{{escape $app.Cmd}}</div>
<div style="display:flex;gap:0.3rem;flex-wrap:wrap">
{{if eq $app.Status "运行中"}}
<button class="btn btn-warning btn-sm" onclick="postAction('/stop/{{$name}}')"><i class="fa-solid fa-stop"></i> 停止</button>
<button class="btn btn-primary btn-sm" onclick="postAction('/restart/{{$name}}')"><i class="fa-solid fa-rotate"></i> 重启</button>
{{else}}
<button class="btn btn-success btn-sm" onclick="postAction('/start/{{$name}}')"><i class="fa-solid fa-play"></i> 启动</button>
{{end}}
<a href="/edit/{{$name}}" class="btn btn-ghost btn-sm"><i class="fa-solid fa-pen"></i> 编辑</a>
<a href="/log/{{$name}}" class="btn btn-ghost btn-sm"><i class="fa-solid fa-file-lines"></i> 日志</a>
{{if $app.URL}}<a href="{{$app.URL}}" target="_blank" class="btn btn-ghost btn-sm"><i class="fa-solid fa-globe"></i> 网页</a>{{end}}
<button class="btn btn-danger btn-sm" onclick="if(confirm('确定删除 {{$name}}？'))postAction('/delete/{{$name}}')"><i class="fa-solid fa-trash"></i> 删除</button>
</div>
</div>
</div>
{{end}}
</div>
</div>
</div>
</div>
` + layoutJS + `
<script>
var selectMode=false;
var appSearch=document.getElementById('appSearch');
if(appSearch){
appSearch.addEventListener('input',function(){
var q=this.value.toLowerCase();
var items=document.querySelectorAll('.app-item');
for(var i=0;i<items.length;i++){
var n=items[i].getAttribute('data-name')||'';
var c=items[i].getAttribute('data-cmd')||'';
if(q&&n.indexOf(q)<0&&c.indexOf(q)<0){items[i].style.display='none';}
else{items[i].style.display='';}
}
});
}
function toggleSelectMode(){
selectMode=!selectMode;
var cbs=document.querySelectorAll('.app-cb');
var labels=document.querySelectorAll('.app-name.app-cb');
var batchBar=document.getElementById('batchBar');
var selectBtn=document.getElementById('selectAllBtn');
if(selectMode){
for(var i=0;i<cbs.length;i++){cbs[i].checked=false;}
for(var i=0;i<labels.length;i++){labels[i].style.display='';}
batchBar.classList.add('active');
selectBtn.innerHTML='<i class="fa-solid fa-check-double"></i> 全选';
}else{
for(var i=0;i<labels.length;i++){labels[i].style.display='none';}
batchBar.classList.remove('active');
updateSelCount();
}
}
function updateSelCount(){
var cbs=document.querySelectorAll('.app-cb:checked');
var cnt=cbs.length;
var el=document.getElementById('selCount');
if(el)el.textContent='已选 '+cnt;
}
document.addEventListener('change',function(e){if(e.target&&e.target.classList.contains('app-cb'))updateSelCount();});
function postAction(url){
var csrfEl=document.querySelector('input[name="csrf_token"]');
var csrf=csrfEl?csrfEl.value:'';
fetch(url,{method:'POST',headers:csrf?{'X-CSRF-Token':csrf,'Content-Type':'application/x-www-form-urlencoded'}:{'Content-Type':'application/x-www-form-urlencoded'},body:'csrf_token='+csrf}).then(function(){location.reload()}).catch(function(){alert('请求失败');});
}
function batchAction(action){
var cbs=document.querySelectorAll('.app-cb:checked');
if(cbs.length===0)return;
var names=[];
for(var i=0;i<cbs.length;i++){names.push(cbs[i].getAttribute('data-name'));}
var msg='确定';
if(action==='start')msg='确定启动: ';
else if(action==='stop')msg='确定停止: ';
else if(action==='delete')msg='确定删除: ';
if(!confirm(msg+names.join(', ')))return;
var csrfEl=document.querySelector('input[name="csrf_token"]');
var csrf=csrfEl?csrfEl.value:'';
fetch('/create/batch',{method:'POST',headers:{'Content-Type':'application/x-www-form-urlencoded'},body:'action='+action+'&names='+encodeURIComponent(names.join(','))+'&csrf_token='+csrf}).then(function(r){return r.json()}).then(function(data){
if(data.error){alert(data.error);}
else{location.reload();}
}).catch(function(){alert('请求失败');});
}
</script>
</body>
</html>`
