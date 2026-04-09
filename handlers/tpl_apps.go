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
</style>
</head>
<body>
<div class="bg-layer"{{if getBgUrl}} style="background-image:url('{{getBgUrl}}')"{{end}}></div>
{{.Sidebar}}
<div class="main">
{{.Topbar}}
<div class="content">
<div class="glass" style="padding:1rem;margin-bottom:1rem">
<div style="display:flex;gap:0.5rem;margin-bottom:1rem">
<button type="button" class="btn btn-primary btn-sm" id="tabDownload" onclick="switchTab('download')">下载安装</button>
<button type="button" class="btn btn-ghost btn-sm" id="tabManual" onclick="switchTab('manual')">手动添加</button>
</div>
<form id="createForm" style="display:grid;gap:0.8rem">
<div id="downloadPanel">
<div style="display:grid;grid-template-columns:1fr 1fr;gap:0.8rem">
<input name="name" placeholder="应用名称" required class="input" id="appName">
<input name="cmd" placeholder="启动命令（留空自动检测）" class="input" id="appCmd">
</div>
<input name="url" placeholder="下载地址（必填）" required class="input" id="appUrl" style="margin-top:0.3rem">
<input name="setup_cmd" placeholder="首次运行命令（可选）" class="input" id="appSetup" style="margin-top:0.3rem">
<div style="display:grid;grid-template-columns:1fr 1fr;gap:0.8rem;margin-top:0.3rem">
<label class="check-label" style="justify-content:center"><input type="checkbox" name="auto_extract" id="autoExtract"><div class="check-box"><i class="fa-solid fa-check"></i></div><span style="color:var(--text2);font-size:0.72rem">自动解压</span></label>
<label class="check-label" style="justify-content:center"><input type="checkbox" name="make_exec" id="makeExec"><div class="check-box"><i class="fa-solid fa-check"></i></div><span style="color:var(--text2);font-size:0.72rem">赋予权限</span></label>
</div>
</div>
<div id="manualPanel" style="display:none">
<div style="font-size:0.65rem;color:var(--text2);margin-bottom:0.5rem;padding:0.4rem;background:var(--card);border-radius:6px"><i class="fa-solid fa-circle-info" style="margin-right:0.3rem"></i>未找到服务？请先查找文件路径和进程名（系统监控页面可查看运行中进程）</div>
<div style="display:grid;grid-template-columns:1fr 1fr;gap:0.8rem">
<input name="manual_name" placeholder="应用名称" class="input" id="manualName">
<input name="manual_path" placeholder="应用目录路径" class="input" id="manualPath">
</div>
<input name="manual_cmd" placeholder="启动命令" class="input" id="manualCmd" style="margin-top:0.3rem">
<input name="manual_workdir" placeholder="工作目录（可选）" class="input" id="manualWorkDir" style="margin-top:0.3rem">
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
<button class="btn btn-ghost btn-sm" id="selectAllBtn" title="全选"><i class="fa-solid fa-check-double"></i> 全选</button>
<button class="btn btn-success btn-sm" id="batchStartBtn" title="批量启动" style="display:none"><i class="fa-solid fa-play"></i> 启动</button>
<button class="btn btn-warning btn-sm" id="batchStopBtn" title="批量停止" style="display:none"><i class="fa-solid fa-stop"></i> 停止</button>
<button class="btn btn-danger btn-sm" id="batchDeleteBtn" title="批量删除" style="display:none"><i class="fa-solid fa-trash"></i> 删除</button>
<button class="btn btn-ghost btn-sm" id="cancelSelectBtn" title="取消选择" style="display:none"><i class="fa-solid fa-xmark"></i> 取消</button>
<span id="selCount" style="font-size:0.7rem;color:var(--text2);display:none">已选 0</span>
</div>
</div>
<div class="app-grid">
{{range $name, $app := .Apps}}
<div class="card app-item" data-name="{{tolower $name}}" data-cmd="{{tolower $app.Cmd}}">
<label style="display:flex;align-items:center;gap:0.5rem;flex:1;min-width:0;cursor:pointer">
<input type="checkbox" class="app-cb" data-name="{{$name}}" style="accent-color:#e53e3e;flex-shrink:0">
<div style="flex:1;min-width:0">
<div style="display:flex;align-items:center;gap:0.5rem;flex-wrap:wrap">
<span class="app-name">{{$name}}</span>
{{if eq $app.Status "运行中"}}<span class="badge badge-running"><span style="width:5px;height:5px;background:#34d399;border-radius:50%"></span>运行中</span>{{else}}<span class="badge badge-stopped"><span style="width:5px;height:5px;background:#f87171;border-radius:50%"></span>已停止</span>{{end}}
{{if $app.AutoStart}}<span class="badge" style="background:rgba(59,130,246,0.1);color:#60a5fa;border:1px solid rgba(59,130,246,0.2)"><i class="fa-solid fa-rotate" style="margin-right:0.2rem"></i>自启</span>{{end}}
{{if $app.Version}}<span class="badge" style="background:rgba(229,62,62,0.1);color:#fc8181;border:1px solid rgba(229,62,62,0.2)">v{{$app.Version}}</span>{{end}}
</div>
<div class="app-cmd"><i class="fa-solid fa-terminal" style="margin-right:0.2rem"></i>{{escape $app.Cmd}}</div>
</div>
</label>
<div style="display:flex;gap:0.3rem;flex-wrap:wrap;flex-shrink:0">
{{if eq $app.Status "运行中"}}
<form action="/stop/{{$name}}" method="post"><button class="btn btn-warning btn-sm"><i class="fa-solid fa-stop"></i> 停止</button></form>
<form action="/restart/{{$name}}" method="post"><button class="btn btn-primary btn-sm"><i class="fa-solid fa-rotate"></i> 重启</button></form>
{{else}}
<form action="/start/{{$name}}" method="post"><button class="btn btn-success btn-sm"><i class="fa-solid fa-play"></i> 启动</button></form>
{{end}}
<a href="/edit/{{$name}}" class="btn btn-ghost btn-sm"><i class="fa-solid fa-pen"></i> 编辑</a>
<a href="/log/{{$name}}" class="btn btn-ghost btn-sm"><i class="fa-solid fa-file-lines"></i> 日志</a>
{{if $app.URL}}<a href="{{$app.URL}}" target="_blank" class="btn btn-ghost btn-sm" title="打开网页"><i class="fa-solid fa-globe"></i> 网页</a>{{end}}
<form action="/toggle-auto/{{$name}}" method="post"><button class="btn btn-ghost btn-sm"><i class="fa-solid fa-rotate"></i> 自启</button></form>
<form action="/delete/{{$name}}" method="post" onsubmit="return confirm('确定删除 {{$name}}？')"><button class="btn btn-danger btn-sm"><i class="fa-solid fa-trash"></i> 删除</button></form>
</div>
</div>
{{end}}
</div>
</div>
</div>
</div>
` + layoutJS + `
<script>
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
var selAll=document.getElementById('selectAllBtn');
var cancelBtn=document.getElementById('cancelSelectBtn');
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
var sa=document.getElementById('selectAllBtn');
if(el){el.textContent='已选 '+cnt;el.style.display=cnt>0?'':'none';}
if(sb)sb.style.display=cnt>0?'':'none';
if(st)st.style.display=cnt>0?'':'none';
if(sd)sd.style.display=cnt>0?'':'none';
if(sa)sa.style.display=cnt>0?'none':'';
if(cancelBtn)cancelBtn.style.display=cnt>0?'':'none';
}
if(cancelBtn){cancelBtn.addEventListener('click',function(){
var cbs=document.querySelectorAll('.app-cb');
for(var i=0;i<cbs.length;i++){cbs[i].checked=false;}
updateSelCount();
});}
document.addEventListener('change',function(e){if(e.target&&e.target.classList.contains('app-cb'))updateSelCount();});
function batchAction(action,confirmMsg){
var cbs=document.querySelectorAll('.app-cb:checked');
if(cbs.length===0)return;
var names=[];
for(var i=0;i<cbs.length;i++){names.push(cbs[i].getAttribute('data-name'));}
if(!confirm(confirmMsg+names.join(', ')))return;
var csrfEl=document.querySelector('input[name="csrf_token"]');
var csrf=csrfEl?csrfEl.value:'';
fetch('/create/batch',{method:'POST',headers:{'Content-Type':'application/x-www-form-urlencoded'},body:'action='+action+'&names='+encodeURIComponent(names.join(','))+'&csrf_token='+csrf}).then(function(r){return r.json()}).then(function(data){
if(data.error){alert(data.error);}
else{location.reload();}
}).catch(function(){alert('请求失败');});
}
var bst=document.getElementById('batchStartBtn');
if(bst)bst.addEventListener('click',function(){batchAction('start','确定启动: ')});
var bsp=document.getElementById('batchStopBtn');
if(bsp)bsp.addEventListener('click',function(){batchAction('stop','确定停止: ')});
var bdl=document.getElementById('batchDeleteBtn');
if(bdl)bdl.addEventListener('click',function(){batchAction('delete','确定删除: ')});
var tabDl=document.getElementById('tabDownload');
var tabMl=document.getElementById('tabManual');
var dl=document.getElementById('downloadPanel');
var ml=document.getElementById('manualPanel');
var createBtn=document.getElementById('createBtn');
function switchTab(tab){
if(tab==='download'){
dl.style.display='block';ml.style.display='none';
if(tabDl){tabDl.className='btn btn-primary btn-sm';}
if(tabMl){tabMl.className='btn btn-ghost btn-sm';}
if(createBtn){createBtn.innerHTML='<i class="fa-solid fa-plus"></i>创建应用';}
}else{
dl.style.display='none';ml.style.display='block';
if(tabDl){tabDl.className='btn btn-ghost btn-sm';}
if(tabMl){tabMl.className='btn btn-primary btn-sm';}
if(createBtn){createBtn.innerHTML='<i class="fa-solid fa-plus"></i>添加应用';}
}
}
var form=document.getElementById('createForm');
if(form){
form.addEventListener('submit',function(e){
e.preventDefault();
var isManual=document.getElementById('manualPanel').style.display==='block';
var nameEl=isManual?document.getElementById('manualName'):document.getElementById('appName');
var prog=document.getElementById('createProgress');
var fill=document.getElementById('progressFill');
var text=document.getElementById('progressText');
if(!nameEl.value.trim()){alert('请输入应用名称');return;}
prog.style.display='block';fill.style.width='10%';text.textContent=isManual?'添加应用...':'准备创建...';
var fd=new FormData();
if(isManual){
fd.append('name',document.getElementById('manualName').value.trim());
fd.append('path',document.getElementById('manualPath').value.trim());
fd.append('cmd',document.getElementById('manualCmd').value.trim());
fd.append('workdir',document.getElementById('manualWorkDir').value.trim());
fd.append('url',document.getElementById('manualUrl').value.trim());
fd.append('auto',document.getElementById('manualAuto').checked?'1':'0');
var progUrl='/create/manual';
fetch(progUrl,{method:'POST',body:fd}).then(function(r){return r.json()}).then(function(data){
fill.style.width='100%';text.textContent=data.error||'完成';if(!data.error){setTimeout(function(){location.reload();},500);}else{prog.style.display='none';alert(data.error);}}
}).catch(function(e){prog.style.display='none';alert('请求失败: '+e);});
}else{
fd.append('name',document.getElementById('appName').value.trim());
fd.append('cmd',document.getElementById('appCmd').value.trim());
fd.append('url',document.getElementById('appUrl').value.trim());
fd.append('setup_cmd',document.getElementById('appSetup').value.trim());
fd.append('auto_extract',document.getElementById('autoExtract').checked?'on':'');
fd.append('make_exec',document.getElementById('makeExec').checked?'on':'');
fetch('/create',{method:'POST',body:fd}).then(function(r){return r.json()}).then(function(data){
if(data.error){fill.style.width='0%';prog.style.display='none';alert(data.error);return;}
fill.style.width='30%';text.textContent='正在下载...';
var taskId=data.id;
var pollInt=setInterval(function(){
fetch('/create/progress/'+taskId).then(function(r){return r.json()}).then(function(t){
if(t&&t.status){
fill.style.width=(t.progress||0)+'%';text.textContent=t.message||t.status;
if(t.status==='completed'||t.status==='error'){clearInterval(pollInt);if(t.status==='completed'){location.reload();}else{prog.style.display='none';alert(t.message);}}
}
}).catch(function(){clearInterval(pollInt);});
},500);
}).catch(function(e){prog.style.display='none';alert('请求失败');});
}
});
}
</script>
</body>
</html>`
