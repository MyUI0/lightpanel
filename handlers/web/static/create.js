
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
[downloadPanel,manualPanel].forEach(function(panel){ panel.querySelectorAll('input,select,textarea').forEach(function(input){input.disabled=(panel.id==='manualPanel')!==(tab==='manual')}); });
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
switchTab('download');
var pollId=null;
var activeTaskId=LP.storage.get('lp_createTask');
var activeTaskErr=LP.storage.get('lp_createErr');
function showBanner(){if(banner)banner.style.display='block'}
function hideBanner(){if(banner)banner.style.display='none';LP.storage.remove('lp_createTask');LP.storage.remove('lp_createErr')}
function pollProgress(id){
LP.request('/create/progress/'+id).then(function(r){return r.json()}).then(function(t){
if(!t||!t.status)return;
if(t.status==='creating'||t.status==='downloading'||t.status==='paused'||t.status==='deploying'){
var pct=t.progress||0;
if(fill)fill.style.width=pct+'%';
if(text)text.textContent=t.status==='paused'?'已暂停，请到下载管理继续':(t.message||'处理中... '+pct+'%');
if(bannerMsg)bannerMsg.textContent=t.message||'正在创建...';
if(bannerFill)bannerFill.style.width=pct+'%';
showBanner();pollId=setTimeout(function(){pollProgress(id)},500);
}else if(t.status==='completed'){
if(fill)fill.style.width='100%';if(text)text.textContent='创建完成！';
if(bannerFill)bannerFill.style.width='100%';if(bannerMsg)bannerMsg.textContent='创建完成！';
showBanner();clearTimeout(pollId);hideBanner();setTimeout(function(){location.reload()},600);
}else if(t.status==='error'||t.status==='cancelled'){
if(text){text.textContent='错误: '+t.message;text.style.color='#f87171'}
if(bannerMsg)bannerMsg.textContent='创建失败: '+t.message;
if(bannerFill)bannerFill.style.width='100%';
if(banner)banner.style.borderColor='rgba(239,68,68,0.3)';
showBanner();btn.disabled=false;btn.innerHTML='<i class="fa-solid fa-plus"></i>创建应用';
clearTimeout(pollId);LP.storage.set('lp_createErr','1');LP.storage.remove('lp_createTask');
}else if(t.status==='not_found'){hideBanner()}
}).catch(function(){pollId=setTimeout(function(){pollProgress(id)},1000)});
}
if(activeTaskErr){
if(banner){banner.style.display='block';banner.style.borderColor='rgba(239,68,68,0.3)';bannerMsg.textContent='上次创建失败';bannerFill.style.width='100%'}
LP.storage.remove('lp_createErr');
}
if(activeTaskId){pollProgress(activeTaskId)}
form.addEventListener('submit',function(e){
e.preventDefault();
var isManual=manualPanel.style.display==='block';
var nameEl=isManual?document.getElementById('manualName'):document.getElementById('appName');
if(!nameEl)return;
var name=nameEl.value.trim();
if(!name){nameEl.focus();return}
LP.storage.remove('lp_createErr');
btn.disabled=true;btn.innerHTML='<i class="fa-solid fa-spinner fa-spin"></i>处理中...';
prog.style.display='block';fill.style.width='10%';text.textContent=isManual?'添加应用...':'准备创建...';
if(isManual){
var mName=document.getElementById('manualName').value.trim();
var mPath=document.getElementById('manualPath').value.trim();
var mCmd=document.getElementById('manualCmd').value.trim();
var mWorkDir=document.getElementById('manualWorkDir').value.trim();
var mIcon=document.getElementById('manualIcon').value.trim();
var mUrl=document.getElementById('manualUrl').value.trim();
var mAuto=document.getElementById('manualAuto').checked;
if(!mPath||!mCmd){text.textContent='路径和命令必填';text.style.color='#f87171';btn.disabled=false;btn.innerHTML='<i class="fa-solid fa-plus"></i>创建应用';return}
var fd=new FormData();
fd.append('name',mName);fd.append('path',mPath);fd.append('cmd',mCmd);
if(mWorkDir)fd.append('work_dir',mWorkDir);
if(mIcon)fd.append('icon',mIcon);
if(mUrl)fd.append('url',mUrl);
if(mAuto)fd.append('auto','on');
LP.request('/create/manual',{method:'POST',body:fd}).then(function(r){return r.json()}).then(function(data){
if(data.ok){text.textContent='添加成功！';fill.style.width='100%';setTimeout(function(){location.reload()},500)}
else{text.textContent=data.error||'添加失败';text.style.color='#f87171';btn.disabled=false;btn.innerHTML='<i class="fa-solid fa-plus"></i>创建应用'}
}).catch(function(){text.textContent='网络错误';btn.disabled=false;btn.innerHTML='<i class="fa-solid fa-plus"></i>创建应用'});
return;
}
var fd=new FormData(form);
LP.request('/create',{method:'POST',body:fd}).then(function(r){
var ct=r.headers.get('content-type')||'';
if(ct.indexOf('json')===-1){text.textContent='请求失败';btn.disabled=false;btn.innerHTML='<i class="fa-solid fa-plus"></i>创建应用';return}
return r.json()}).then(function(data){
if(!data)return;
if(data.ok&&data.task){LP.storage.set('lp_createTask',data.task);pollProgress(data.task);setTimeout(function(){window.location.replace(data.redirect||'/')},300)}
else if(data.error){text.textContent=data.error;text.style.color='#f87171';if(bannerMsg)bannerMsg.textContent='创建失败: '+data.error;if(banner)banner.style.borderColor='rgba(239,68,68,0.3)';showBanner();LP.storage.set('lp_createErr','1');btn.disabled=false;btn.innerHTML='<i class="fa-solid fa-plus"></i>创建应用'}
}).catch(function(){text.textContent='网络错误';btn.disabled=false;btn.innerHTML='<i class="fa-solid fa-plus"></i>创建应用'});
});
})();
