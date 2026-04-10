package handlers

var htmlIndex = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width,initial-scale=1.0">
<title>朱雀面板</title>
<style>
` + layoutCSS + `
.app-grid{display:grid;grid-template-columns:repeat(2,1fr);gap:0.6rem}
@media(max-width:800px){.app-grid{grid-template-columns:1fr}}
.app-grid{display:grid;grid-template-columns:repeat(2,1fr);gap:0.6rem}
@media(max-width:800px){.app-grid{grid-template-columns:1fr}}
.app-item{display:flex;gap:0.6rem;padding:0.8rem;min-height:90px;position:relative}
.app-item .app-icon{width:44px;height:44px;border-radius:12px;background:rgba(229,62,62,0.15);display:flex;align-items:center;justify-content:center;flex-shrink:0;overflow:hidden}
.app-item .app-icon img{width:100%;height:100%;object-fit:cover;border-radius:12px}
.app-item .app-icon i{font-size:1.2rem;color:#e53e3e}
.app-item .app-info{flex:1;min-width:0;display:flex;flex-direction:column;justify-content:center}
.app-item .app-name-row{display:flex;align-items:center;gap:0.4rem;flex-wrap:wrap}
.app-item .app-name{font-weight:600;color:var(--text);font-size:0.88rem}
.app-item .app-cmd{font-size:0.7rem;color:var(--text2);margin-top:0.2rem;max-width:280px;overflow:hidden;text-overflow:ellipsis;white-space:nowrap}
.app-item .app-actions{display:flex;gap:0.2rem;flex-wrap:wrap;padding:0.5rem 0 0;margin:0.4rem 0 0}
.app-item .app-actions .btn{padding:0.28rem 0.5rem;font-size:0.64rem;border-radius:6px}
.app-item .app-cb{width:16px;height:16px;accent-color:#e53e3e;position:absolute;left:0.6rem;top:0.5rem;opacity:0;transition:opacity 0.15s}
.app-item.selected .app-cb,.app-item:has(.app-cb:checked) .app-cb{opacity:1}
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
<div id="activeTaskBanner" style="display:none;background:rgba(229,62,62,0.08);border:1px solid rgba(229,62,62,0.2);border-radius:10px;padding:0.7rem 0.8rem;margin-bottom:0.8rem">
<div style="font-size:0.8rem;color:#fc8181;margin-bottom:0.4rem"><i class="fa-solid fa-spinner fa-spin" style="margin-right:0.3rem"></i><span id="activeTaskMsg">正在创建...</span></div>
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
<input type="checkbox" class="app-cb" data-name="{{$name}}" id="cb-{{$name}}">
<label style="display:flex;gap:0.6rem;cursor:pointer;flex:1" for="cb-{{$name}}">
{{if $app.Icon}}
<div class="app-icon"><img src="{{$app.Icon}}" onerror="this.parentElement.innerHTML='<i class=\\'fa-solid fa-box\\'></i>'"></div>
{{else}}
<div class="app-icon"><i class="fa-solid fa-box"></i></div>
{{end}}
<div class="app-info">
<div class="app-name-row">
<span class="app-name">{{$name}}</span>
{{if eq $app.Status "运行中"}}<span class="badge badge-running"><span style="width:5px;height:5px;background:#34d399;border-radius:50%"></span>运行中</span>{{else}}<span class="badge badge-stopped"><span style="width:5px;height:5px;background:#f87171;border-radius:50%"></span>已停止</span>{{end}}
{{if $app.AutoStart}}<span class="badge" style="background:rgba(59,130,246,0.1);color:#60a5fa;border:1px solid rgba(59,130,246,0.2)"><i class="fa-solid fa-power-off" style="margin-right:0.2rem"></i>自启</span>{{end}}
{{if $app.Version}}<span class="badge" style="background:rgba(229,62,62,0.1);color:#fc8181;border:1px solid rgba(229,62,62,0.2)">v{{$app.Version}}</span>{{end}}
</div>
<div class="app-cmd"><i class="fa-solid fa-terminal" style="margin-right:0.2rem"></i>{{escape $app.Cmd}}</div>
</div>
</label>
<div class="app-actions">
{{if eq $app.Status "运行中"}}
<form action="/stop/{{$name}}" method="post"><button class="btn btn-warning"><i class="fa-solid fa-stop"></i>停止</button></form>
<form action="/restart/{{$name}}" method="post"><button class="btn btn-primary"><i class="fa-solid fa-rotate"></i>重启</button></form>
{{else}}
<form action="/start/{{$name}}" method="post"><button class="btn btn-success"><i class="fa-solid fa-play"></i>启动</button></form>
{{end}}
{{if not $app.AutoStart}}
<form action="/toggle-auto/{{$name}}" method="post"><button class="btn btn-ghost" title="开启自启"><i class="fa-solid fa-power-off"></i>自启</button></form>
{{else}}
<form action="/toggle-auto/{{$name}}" method="post"><button class="btn" style="background:rgba(59,130,246,0.15);color:#60a5fa;border:none" title="关闭自启"><i class="fa-solid fa-power-off"></i>自启</button></form>
{{end}}
<a href="/edit/{{$name}}" class="btn btn-ghost"><i class="fa-solid fa-pen"></i>编辑</a>
<a href="/log/{{$name}}" class="btn btn-ghost"><i class="fa-solid fa-file-lines"></i>日志</a>
{{if $app.URL}}<a href="{{$app.URL}}" target="_blank" class="btn btn-ghost" title="打开网页"><i class="fa-solid fa-globe"></i>网页</a>{{end}}
<form action="/delete/{{$name}}" method="post" onsubmit="return confirm('确定删除 {{$name}}？')"><button class="btn btn-danger"><i class="fa-solid fa-trash"></i>删除</button></form>
</div>
</div>
</div>
{{end}}
</div>
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
var appItems=document.querySelectorAll('.app-item');
if(appItems.length>0){
fetch('/api/updates').then(function(r){return r.json()}).then(function(updates){
if(!updates)return;
for(var name in updates){
var el=document.querySelector('.app-item[data-name="'+name.toLowerCase()+'"]');
if(el){
var badge=document.createElement('span');
badge.className='badge';
badge.style.cssText='background:rgba(245,158,11,0.2);color:#fbbf24;border:1px solid rgba(245,158,11,0.3)';
badge.innerHTML='<i class="fa-solid fa-arrow-up" style="margin-right:0.2rem"></i>有更新 v'+updates[name];
var badgeArea=el.querySelector('[style*="flex-wrap:wrap"]');
if(badgeArea)badgeArea.appendChild(badge);
}
}
}).catch(function(){});
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
var cancelBtn=document.getElementById('cancelSelectBtn');
if(selAll){
selAll.addEventListener('click',function(){
var cbs=document.querySelectorAll('.app-cb');
var allChecked=true;
for(var i=0;i<cbs.length;i++){if(!cbs[i].checked){allChecked=false;break;}}
for(var i=0;i<cbs.length;i++){cbs[i].checked=!allChecked;}
var items=document.querySelectorAll('.app-item');
for(var i=0;i<items.length;i++){if(!allChecked){items[i].classList.add('selected');}else{items[i].classList.remove('selected');}}
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
var items=document.querySelectorAll('.app-item');
if(cnt>0){
for(var i=0;i<items.length;i++){items[i].classList.add('selected');}
}else{
for(var i=0;i<items.length;i++){items[i].classList.remove('selected');}
}
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
document.addEventListener('change',function(e){
if(e.target&&e.target.classList.contains('app-cb')){
var items=document.querySelectorAll('.app-item');
for(var i=0;i<items.length;i++){
var cb=items[i].querySelector('.app-cb');
if(cb&&cb.checked){items[i].classList.add('selected');}else{items[i].classList.remove('selected');}
}
updateSelCount();}
});
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
