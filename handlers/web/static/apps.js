(() => {
 'use strict';
 const $=id=>document.getElementById(id), items=()=>[...document.querySelectorAll('.app-item')];
 let filter='all', busy=false, refreshing=false;
 const dialog=$('createDialog');
 document.querySelectorAll('[data-open-create]').forEach(b=>b.addEventListener('click',()=>dialog.showModal()));
 $('closeCreate').addEventListener('click',()=>dialog.close());
 dialog.addEventListener('click',e=>{if(e.target===dialog){const r=dialog.getBoundingClientRect();if(e.clientX<r.left||e.clientX>r.right||e.clientY<r.top||e.clientY>r.bottom)dialog.close();}});
 function selection(){
  let count=0;
  items().forEach(item=>{const checked=item.querySelector('.app-cb').checked;item.classList.toggle('selected',checked);if(checked)count++;});
  $('batchBar').hidden=count===0; $('selCount').textContent='已选 '+count+' 个应用';
 }
 function applyFilters(){
  const q=$('appSearch').value.trim().toLowerCase(); let visible=0,running=0;
  items().forEach(item=>{
   if(item.dataset.status==='running')running++;
   const matches=(filter==='all'||item.dataset.status===filter)&&(item.dataset.name+' '+item.dataset.cmd).toLowerCase().includes(q);
   item.hidden=!matches;
   if(matches)visible++;else item.querySelector('.app-cb').checked=false;
  });
  $('countAll').textContent=items().length; $('countRunning').textContent=running; $('countStopped').textContent=items().length-running;
  $('visibleCount').textContent='显示 '+visible+' / '+items().length+' 个应用';
  $('noResults').hidden=visible>0||items().length===0; $('emptyState').hidden=items().length>0;
  selection();
 }
 document.querySelectorAll('[data-filter]').forEach(button=>button.addEventListener('click',()=>{
  filter=button.dataset.filter;
  document.querySelectorAll('[data-filter]').forEach(b=>b.setAttribute('aria-selected',String(b===button))); applyFilters();
 }));
 $('appSearch').addEventListener('input',applyFilters);
 $('clearFilters').addEventListener('click',()=>{$('appSearch').value='';document.querySelector('[data-filter=all]').click();});
 $('selectAllBtn').addEventListener('click',()=>{const visible=items().filter(i=>!i.hidden);const checked=visible.length>0&&visible.every(i=>i.querySelector('.app-cb').checked);visible.forEach(i=>i.querySelector('.app-cb').checked=!checked);selection();});
 $('cancelSelectBtn').addEventListener('click',()=>{items().forEach(i=>i.querySelector('.app-cb').checked=false);selection();});
 document.addEventListener('change',e=>{if(e.target.matches('.app-cb'))selection();});
 async function refresh(){
  if(refreshing)return;refreshing=true;
  try {
   const data=await(await LP.request('/api/overview')).json();
   document.querySelectorAll('[data-metric]').forEach(el=>el.textContent=data[el.dataset.metric]);
   document.querySelectorAll('[data-meter]').forEach(el=>el.style.width=Math.max(0,Math.min(100,data[el.dataset.meter]))+'%');
   items().forEach(item=>{
    const state=data.apps[item.dataset.name];if(!state){item.remove();return;}
    const running=state.status==='运行中';item.dataset.status=running?'running':'stopped';
    const badge=item.querySelector('[data-app-status]');badge.textContent=state.status;badge.className='badge '+(running?'badge-running':'badge-stopped');
    const button=item.querySelector('[data-toggle-state]');button.dataset.action=running?'stop':'start';button.textContent=running?'停止':'启动';button.className='btn btn-sm '+(running?'btn-warning':'btn-success');
   });
   $('refreshStatus').textContent='更新于 '+new Date().toLocaleTimeString('zh-CN',{hour12:false});applyFilters();
  }catch(e){$('refreshStatus').textContent='连接中断，稍后重试';throw e;}finally{refreshing=false;}
 }
 async function action(name, action){
  const response=await LP.request('/'+action+'/'+encodeURIComponent(name),{method:'POST',headers:{Accept:'application/json'}});
  const result=await response.json();if(!result.ok)throw new Error(result.error||'操作失败');
 }
 document.addEventListener('click',async e=>{
  const button=e.target.closest('[data-action]');if(!button||busy)return;
  const item=button.closest('.app-item'),name=item.dataset.name,kind=button.dataset.action;
  if(kind==='delete'&&!await LP.confirm('删除应用“'+name+'”？应用目录内的文件也会被删除，此操作不可撤销。'))return;
  busy=true;button.disabled=true;
  try{
   await action(name,kind);
   if(kind==='delete')item.remove();
   if(kind==='toggle-auto'){location.reload();return;}
   LP.toast(kind==='start'||kind==='restart'?'启动请求已执行，请查看实时状态':'操作完成');await refresh();
  }catch(error){LP.toast(error.message,true);}finally{busy=false;button.disabled=false;}
 });
 document.querySelectorAll('[data-batch]').forEach(button=>button.addEventListener('click',async()=>{
  if(busy)return;
  const chosen=items().filter(i=>i.querySelector('.app-cb').checked),kind=button.dataset.batch;
  const verb={start:'启动',stop:'停止',delete:'删除'}[kind];
  if(!await LP.confirm('确定'+verb+'选中的 '+chosen.length+' 个应用？'+(kind==='delete'?'应用目录内的文件也会被删除，此操作不可撤销。':'')))return;
  busy=true;const controls=[...document.querySelectorAll('[data-batch],[data-action],.app-cb,#cancelSelectBtn,#selectAllBtn')];controls.forEach(b=>b.disabled=true);
  let completed=0;const failures=[];
  // Sequential requests ensure every result is accounted for before navigation.
  for(const item of chosen){
   try{await action(item.dataset.name,kind);completed++;item.querySelector('.app-cb').checked=false;if(kind==='delete')item.remove();}
   catch(e){failures.push(item.dataset.name+'：'+e.message);}
   $('selCount').textContent='已处理 '+(completed+failures.length)+' / '+chosen.length;
  }
  busy=false;controls.forEach(b=>b.disabled=false);selection();
  LP.toast(verb+'完成 '+completed+' 个'+(failures.length?'；失败 '+failures.join('；'):''),failures.length>0);
  refresh().catch(()=>{});
 }));
 $('refreshBtn').addEventListener('click',()=>refresh().catch(e=>LP.toast(e.message,true)));
 async function poll(){if(!document.hidden&&!busy)await refresh().catch(()=>{});setTimeout(poll,5000);}
 applyFilters();setTimeout(poll,5000);
 if(items().length) LP.request('/api/updates').then(r=>r.json()).then(updates=>{
  items().forEach(item=>{if(updates?.[item.dataset.name]){const badge=item.querySelector('[data-update]');badge.textContent='可用更新 '+updates[item.dataset.name];badge.hidden=false;}});
 }).catch(()=>{});
 document.addEventListener('visibilitychange',()=>{if(!document.hidden&&!busy)refresh().catch(()=>{});});
})();
