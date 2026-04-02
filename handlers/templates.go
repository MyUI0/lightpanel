package handlers

const htmlIndex = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width,initial-scale=1.0">
<title>LightPanel 管理面板</title>
<script src="https://cdn.tailwindcss.com"></script>
<link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.5.1/css/all.min.css">
<style>
@import url('https://fonts.googleapis.com/css2?family=Inter:wght@300;400;500;600;700&display=swap');
:root { --accent:#6366f1; --accent-glow:rgba(99,102,241,0.4); }
* { margin:0; padding:0; box-sizing:border-box; font-family:'Inter',system-ui,sans-serif; }
body { background:#0f0f23; min-height:100vh; overflow-x:hidden; position:relative; }
.bg-layer { position:fixed; inset:0; z-index:0; background:radial-gradient(ellipse at 20% 50%,rgba(99,102,241,0.15) 0%,transparent 50%),radial-gradient(ellipse at 80% 20%,rgba(168,85,247,0.12) 0%,transparent 50%),radial-gradient(ellipse at 40% 80%,rgba(59,130,246,0.1) 0%,transparent 50%),#0f0f23; }
.bg-layer.bg-img { background-size:cover; background-position:center; background-repeat:no-repeat; }
.bg-layer.bg-img::after { content:''; position:absolute; inset:0; background:linear-gradient(180deg,rgba(15,15,35,0.7) 0%,rgba(15,15,35,0.85) 100%); }
.bg-layer::before { content:''; position:absolute; inset:0; background-image:radial-gradient(rgba(255,255,255,0.03) 1px,transparent 1px); background-size:30px 30px; }
.container { position:relative; z-index:1; max-width:1200px; margin:0 auto; padding:1.5rem; }
.stat-card { background:linear-gradient(135deg,rgba(255,255,255,0.05),rgba(255,255,255,0.02)); backdrop-filter:blur(12px); -webkit-backdrop-filter:blur(12px); border:1px solid rgba(255,255,255,0.08); border-radius:16px; padding:1.25rem; text-align:center; position:relative; overflow:hidden; }
.stat-card::before { content:''; position:absolute; top:0; left:0; right:0; height:3px; background:linear-gradient(90deg,var(--accent),#a855f7); opacity:0.8; }
.stat-card:nth-child(2)::before { background:linear-gradient(90deg,#10b981,#34d399); }
.stat-card:nth-child(3)::before { background:linear-gradient(90deg,#f59e0b,#fbbf24); }
.stat-card:nth-child(4)::before { background:linear-gradient(90deg,#3b82f6,#60a5fa); }
.stat-value { font-size:2rem; font-weight:700; background:linear-gradient(135deg,#fff,rgba(255,255,255,0.7)); -webkit-background-clip:text; -webkit-text-fill-color:transparent; }
.stat-label { font-size:0.75rem; color:rgba(255,255,255,0.5); text-transform:uppercase; letter-spacing:1px; margin-top:0.25rem; }
.btn { display:inline-flex; align-items:center; justify-content:center; gap:0.5rem; padding:0.6rem 1.2rem; border-radius:12px; font-size:0.875rem; font-weight:500; transition:all 0.2s ease; cursor:pointer; border:none; text-decoration:none; }
.btn-primary { background:linear-gradient(135deg,var(--accent),#7c3aed); color:#fff; }
.btn-primary:hover { box-shadow:0 4px 20px var(--accent-glow); transform:translateY(-1px); }
.btn-success { background:linear-gradient(135deg,#10b981,#059669); color:#fff; }
.btn-success:hover { box-shadow:0 4px 20px rgba(16,185,129,0.4); }
.btn-danger { background:linear-gradient(135deg,#ef4444,#dc2626); color:#fff; }
.btn-danger:hover { box-shadow:0 4px 20px rgba(239,68,68,0.4); }
.btn-ghost { background:rgba(255,255,255,0.05); color:rgba(255,255,255,0.7); border:1px solid rgba(255,255,255,0.1); }
.btn-ghost:hover { background:rgba(255,255,255,0.1); color:#fff; }
.btn-sm { padding:0.4rem 0.8rem; font-size:0.75rem; border-radius:8px; }
.btn:disabled { opacity:0.5; cursor:not-allowed; }
.input { width:100%; padding:0.75rem 1rem; border-radius:12px; background:rgba(255,255,255,0.05); border:1px solid rgba(255,255,255,0.1); color:#fff; font-size:0.875rem; transition:all 0.2s ease; }
.input:focus { outline:none; border-color:var(--accent); box-shadow:0 0 0 3px var(--accent-glow); }
.input::placeholder { color:rgba(255,255,255,0.3); }
.badge { display:inline-flex; align-items:center; gap:0.35rem; padding:0.3rem 0.75rem; border-radius:9999px; font-size:0.75rem; font-weight:500; }
.badge-running { background:rgba(16,185,129,0.15); color:#34d399; border:1px solid rgba(16,185,129,0.3); }
.badge-stopped { background:rgba(239,68,68,0.15); color:#f87171; border:1px solid rgba(239,68,68,0.3); }
.badge-crashed { background:rgba(245,158,11,0.15); color:#fbbf24; border:1px solid rgba(245,158,11,0.3); }
.app-grid { display:grid; grid-template-columns:repeat(auto-fill,minmax(340px,1fr)); gap:1rem; }
.app-card { padding:1.25rem; background:rgba(255,255,255,0.03); backdrop-filter:blur(16px); -webkit-backdrop-filter:blur(16px); border:1px solid rgba(255,255,255,0.08); border-radius:16px; transition:all 0.3s cubic-bezier(0.4,0,0.2,1); }
.app-card:hover { background:rgba(255,255,255,0.06); border-color:rgba(255,255,255,0.15); transform:translateY(-2px); box-shadow:0 12px 40px rgba(0,0,0,0.4); }
.app-header { display:flex; justify-content:space-between; align-items:center; margin-bottom:1rem; }
.app-name { font-size:1.1rem; font-weight:600; color:#fff; }
.app-actions { display:grid; grid-template-columns:repeat(3,1fr); gap:0.5rem; margin-top:1rem; }
.app-actions .btn { padding:0.5rem; font-size:0.8rem; }
.app-footer { display:grid; grid-template-columns:1fr 1fr 1fr; gap:0.5rem; margin-top:0.5rem; }
.nav { display:flex; align-items:center; justify-content:space-between; margin-bottom:2rem; }
.nav-brand { display:flex; align-items:center; gap:0.75rem; }
.nav-title { font-size:1.5rem; font-weight:700; background:linear-gradient(135deg,#fff,rgba(255,255,255,0.6)); -webkit-background-clip:text; -webkit-text-fill-color:transparent; }
.nav-links { display:flex; gap:0.5rem; }
.section-title { font-size:1.1rem; font-weight:600; color:#fff; margin-bottom:1rem; display:flex; align-items:center; gap:0.5rem; }
.section-title i { color:var(--accent); }
.glass { background:rgba(255,255,255,0.03); backdrop-filter:blur(20px); -webkit-backdrop-filter:blur(20px); border:1px solid rgba(255,255,255,0.08); border-radius:16px; }
@keyframes pulse-glow { 0%,100%{opacity:1} 50%{opacity:0.5} }
.pulse { animation:pulse-glow 2s infinite; }
.toggle { position:relative; width:40px; height:22px; border-radius:11px; background:rgba(255,255,255,0.15); cursor:pointer; transition:background 0.3s; }
.toggle.on { background:rgba(16,185,129,0.5); }
.toggle::after { content:''; position:absolute; top:2px; left:2px; width:18px; height:18px; border-radius:50%; background:#fff; transition:transform 0.3s; }
.toggle.on::after { transform:translateX(18px); }
@keyframes toast-in { from { transform:translateX(100%); opacity:0; } to { transform:translateX(0); opacity:1; } }
@keyframes toast-out { from { transform:translateX(0); opacity:1; } to { transform:translateX(100%); opacity:0; } }
.toast { position:fixed; top:1.5rem; right:1.5rem; z-index:9999; padding:0.85rem 1.25rem; border-radius:12px; font-size:0.85rem; font-weight:500; color:#fff; max-width:380px; box-shadow:0 8px 32px rgba(0,0,0,0.5); animation:toast-in 0.3s ease; }
.toast-success { background:linear-gradient(135deg,#10b981,#059669); }
.toast-error { background:linear-gradient(135deg,#ef4444,#dc2626); }
.toast-info { background:linear-gradient(135deg,#6366f1,#7c3aed); }
.toast-warning { background:linear-gradient(135deg,#f59e0b,#d97706); }
.overlay { position:fixed; inset:0; z-index:9998; background:rgba(0,0,0,0.6); backdrop-filter:blur(4px); display:flex; align-items:center; justify-content:center; }
.overlay-box { background:#1a1a2e; border:1px solid rgba(255,255,255,0.1); border-radius:20px; padding:2rem; max-width:500px; width:90%; text-align:center; }
@keyframes spin { to { transform:rotate(360deg); } }
.spinner { width:32px; height:32px; border:3px solid rgba(255,255,255,0.1); border-top-color:var(--accent); border-radius:50%; animation:spin 0.8s linear infinite; margin:0 auto 1rem; }
.progress-bar { width:100%; height:6px; background:rgba(255,255,255,0.1); border-radius:3px; overflow:hidden; margin:1rem 0; }
.progress-fill { height:100%; background:linear-gradient(90deg,var(--accent),#a855f7); border-radius:3px; transition:width 0.3s ease; }
.dep-tag { display:inline-block; padding:0.25rem 0.6rem; border-radius:6px; font-size:0.75rem; background:rgba(239,68,68,0.15); color:#f87171; border:1px solid rgba(239,68,68,0.3); margin:0.2rem; }
</style>
</head>
<body>
<div class="bg-layer{{if .BgUrl}} bg-img{{end}}"{{if .BgUrl}} style="background-image:url('{{.BgUrl}}')"{{end}}></div>
<div class="container">
<nav class="nav">
  <div class="nav-brand">
    <div style="width:40px;height:40px;background:linear-gradient(135deg,var(--accent),#a855f7);border-radius:12px;display:flex;align-items:center;justify-content:center;">
      <i class="fa-solid fa-server" style="color:#fff;font-size:1.1rem;"></i>
    </div>
    <div>
      <div class="nav-title">LightPanel</div>
      <div style="font-size:0.75rem;color:rgba(255,255,255,0.4);">{{.Uptime}}</div>
    </div>
  </div>
  <div class="nav-links">
    <a href="/system" class="btn btn-ghost"><i class="fa-solid fa-chart-line"></i>监控</a>
    <a href="/downloads" class="btn btn-ghost"><i class="fa-solid fa-download"></i>下载</a>
    <a href="/store" class="btn btn-ghost"><i class="fa-solid fa-store"></i>商店</a>
    <a href="/setting" class="btn btn-ghost"><i class="fa-solid fa-gear"></i>设置</a>
  </div>
</nav>

<div style="display:grid;grid-template-columns:repeat(4,1fr);gap:1rem;margin-bottom:2rem;">
  <div class="stat-card"><div class="stat-value">{{.Cpu}}%</div><div class="stat-label">CPU 使用率</div></div>
  <div class="stat-card"><div class="stat-value">{{.Mem}}%</div><div class="stat-label">内存使用率</div></div>
  <div class="stat-card"><div class="stat-value">{{.Disk}}%</div><div class="stat-label">磁盘使用率</div></div>
  <div class="stat-card"><div class="stat-value">{{.ProcNum}}</div><div class="stat-label">系统进程</div></div>
</div>

<div class="glass" style="padding:1.5rem;margin-bottom:2rem;">
  <div class="section-title"><i class="fa-solid fa-plus-circle"></i>快速部署</div>
  <form id="createForm" action="/create" method="post" style="display:grid;grid-template-columns:1fr 2fr 1fr auto;gap:0.75rem;align-items:end;">
    <div><label style="font-size:0.75rem;color:rgba(255,255,255,0.5);margin-bottom:0.35rem;display:block;">应用名称</label>
    <input name="name" placeholder="myapp" required class="input"></div>
    <div><label style="font-size:0.75rem;color:rgba(255,255,255,0.5);margin-bottom:0.35rem;display:block;">下载地址 (可选)</label>
    <input name="url" placeholder="https://example.com/app.tar.gz" class="input"></div>
    <div><label style="font-size:0.75rem;color:rgba(255,255,255,0.5);margin-bottom:0.35rem;display:block;">启动命令</label>
    <input name="cmd" placeholder="./start.sh" required class="input"></div>
    <button type="submit" class="btn btn-primary" style="height:44px;"><i class="fa-solid fa-rocket"></i>创建</button>
  </form>
</div>

<div class="section-title"><i class="fa-solid fa-cubes"></i>我的应用 <span style="font-size:0.8rem;color:rgba(255,255,255,0.4);font-weight:400;">({{len .Apps}})</span></div>
<div class="app-grid">
{{range $n,$a := .Apps}}
<div class="app-card" id="card-{{$n}}">
  <div class="app-header">
    <div class="app-name"><i class="fa-solid fa-cube" style="color:var(--accent);margin-right:0.5rem;font-size:0.9rem;"></i>{{$n}}</div>
    <div style="display:flex;align-items:center;gap:0.5rem;">
      <form action="/toggle-auto/{{$n}}" method="post" style="display:inline;" title="开机自启">
        <button type="submit" style="background:none;border:none;padding:0;cursor:pointer;">
          <div class="toggle{{if $a.AutoStart}} on{{end}}"></div>
        </button>
      </form>
      {{if eq $a.Status "运行中"}}
      <span class="badge badge-running"><span class="pulse" style="width:6px;height:6px;background:#34d399;border-radius:50%;"></span>运行中</span>
      {{else}}
      <span class="badge badge-stopped"><span style="width:6px;height:6px;background:#f87171;border-radius:50%;"></span>已停止</span>
      {{end}}
    </div>
  </div>
  <div style="font-size:0.75rem;color:rgba(255,255,255,0.4);margin-bottom:0.75rem;">
    <i class="fa-solid fa-folder" style="margin-right:0.3rem;"></i>{{$a.Path}}
    {{if $a.Created}}<br><i class="fa-solid fa-clock" style="margin-right:0.3rem;"></i>{{$a.Created}}{{end}}
    {{if $a.AutoStart}}<br><i class="fa-solid fa-power-off" style="margin-right:0.3rem;color:#34d399;"></i>开机自启{{end}}
  </div>
  <div class="app-actions">
    <form action="/start/{{$n}}" method="post" style="display:contents;"><button class="btn btn-success app-action-btn" data-name="{{$n}}"><i class="fa-solid fa-play"></i>启动</button></form>
    <form action="/stop/{{$n}}" method="post" style="display:contents;"><button class="btn btn-danger app-action-btn" data-name="{{$n}}"><i class="fa-solid fa-stop"></i>停止</button></form>
    <form action="/restart/{{$n}}" method="post" style="display:contents;"><button class="btn btn-primary app-action-btn" data-name="{{$n}}"><i class="fa-solid fa-rotate"></i>重启</button></form>
  </div>
  <div class="app-footer">
    <a href="/log/{{$n}}" class="btn btn-ghost btn-sm"><i class="fa-solid fa-file-lines"></i>日志</a>
    <a href="/edit/{{$n}}" class="btn btn-ghost btn-sm"><i class="fa-solid fa-pen"></i>编辑</a>
    <form action="/delete/{{$n}}" method="post" style="display:contents;"><button class="btn btn-ghost btn-sm" style="color:#f87171;" onclick="return confirm('确定删除 {{$n}}？\n将清除所有数据！')"><i class="fa-solid fa-trash"></i>删除</button></form>
  </div>
</div>
{{end}}
</div>

{{if eq (len .Apps) 0}}
<div class="glass" style="padding:3rem;text-align:center;margin-top:2rem;">
  <i class="fa-solid fa-inbox" style="font-size:3rem;color:rgba(255,255,255,0.15);margin-bottom:1rem;"></i>
  <p style="color:rgba(255,255,255,0.4);">暂无应用，使用上方表单或应用商店添加</p>
</div>
{{end}}
</div>

<script>
function showToast(msg, type) {
  var t = document.createElement('div');
  t.className = 'toast toast-' + (type || 'info');
  t.textContent = msg;
  document.body.appendChild(t);
  setTimeout(function() { t.style.animation = 'toast-out 0.3s ease forwards'; }, 2500);
  setTimeout(function() { t.remove(); }, 2800);
}

function postForm(form, cb) {
  var btn = form.querySelector('button[type="submit"]');
  if (btn && !btn.disabled) { btn.disabled = true; btn.dataset.orig = btn.innerHTML; btn.innerHTML = '<i class="fa-solid fa-spinner fa-spin"></i>'; }
  var fd = new FormData(form);
  fetch(form.action || form.getAttribute('action'), { method: 'POST', body: fd })
    .then(function(r) { if (r.ok) { showToast('操作成功', 'success'); if (cb) cb(); else setTimeout(function() { location.reload(); }, 800); } else { showToast('操作失败 (' + r.status + ')', 'error'); } })
    .catch(function() { showToast('网络错误', 'error'); })
    .finally(function() { if (btn) { btn.disabled = false; if (btn.dataset.orig) btn.innerHTML = btn.dataset.orig; } });
}

document.getElementById('createForm').addEventListener('submit', function(e) {
  e.preventDefault();
  postForm(this, function() { setTimeout(function() { location.reload(); }, 1000); });
});

document.querySelectorAll('.app-action-btn').forEach(function(btn) {
  btn.closest('form').addEventListener('submit', function(e) {
    e.preventDefault();
    postForm(this, function() { setTimeout(function() { location.reload(); }, 600); });
  });
});

document.querySelectorAll('form[action^="/toggle-auto/"]').forEach(function(f) {
  f.addEventListener('submit', function(e) {
    e.preventDefault();
    var fd = new FormData(this);
    fetch(this.action, { method: 'POST', body: fd }).then(function() { location.reload(); });
  });
});

document.querySelectorAll('form[action^="/delete/"]').forEach(function(f) {
  f.addEventListener('submit', function(e) {
    e.preventDefault();
    if (!confirm('确定删除此应用？\n将清除所有数据！')) return;
    postForm(this);
  });
});

{{if .FailInfo}}
(function() {
  var overlay = document.createElement('div');
  overlay.className = 'overlay';
  overlay.innerHTML = '<div class="overlay-box">' +
    '<div style="font-size:1.1rem;font-weight:600;color:#f87171;margin-bottom:1rem;"><i class="fa-solid fa-triangle-exclamation" style="margin-right:0.5rem;"></i>{{.FailInfo.Name}} 启动失败</div>' +
    '<div style="font-size:0.85rem;color:rgba(255,255,255,0.6);margin-bottom:1rem;text-align:left;background:rgba(0,0,0,0.3);border-radius:10px;padding:0.75rem;max-height:200px;overflow-y:auto;white-space:pre-wrap;font-family:monospace;">{{.FailInfo.Log}}</div>' +
    '{{if .FailInfo.Deps}}<div style="font-size:0.8rem;color:#fbbf24;margin-bottom:1rem;"><i class="fa-solid fa-wrench" style="margin-right:0.5rem;"></i>可能缺失依赖：{{range .FailInfo.Deps}}<span class="dep-tag">{{.}}</span>{{end}}</div>{{end}}' +
    '<button class="btn btn-ghost" onclick="this.closest(\'.overlay\').remove()"><i class="fa-solid fa-xmark" style="margin-right:0.5rem;"></i>关闭</button>' +
    '</div>';
  document.body.appendChild(overlay);
})();
{{end}}
</script>
</body>
</html>`

const htmlEdit = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width,initial-scale=1.0">
<title>编辑应用 - LightPanel</title>
<script src="https://cdn.tailwindcss.com"></script>
<link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.5.1/css/all.min.css">
<style>
@import url('https://fonts.googleapis.com/css2?family=Inter:wght@300;400;500;600;700&display=swap');
* { margin:0; padding:0; box-sizing:border-box; font-family:'Inter',system-ui,sans-serif; }
body { background:#0f0f23; min-height:100vh; position:relative; }
.bg-layer { position:fixed; inset:0; z-index:0; background:radial-gradient(ellipse at 50% 50%,rgba(99,102,241,0.12) 0%,transparent 60%),#0f0f23; }
.container { position:relative; z-index:1; max-width:600px; margin:0 auto; padding:1.5rem; }
.glass { background:rgba(255,255,255,0.03); backdrop-filter:blur(16px); border:1px solid rgba(255,255,255,0.08); border-radius:16px; }
.btn { display:inline-flex; align-items:center; justify-content:center; gap:0.5rem; padding:0.6rem 1.2rem; border-radius:12px; font-size:0.875rem; font-weight:500; transition:all 0.2s ease; cursor:pointer; border:none; text-decoration:none; }
.btn-primary { background:linear-gradient(135deg,#6366f1,#7c3aed); color:#fff; }
.btn-primary:hover { box-shadow:0 4px 20px rgba(99,102,241,0.4); }
.btn-ghost { background:rgba(255,255,255,0.05); color:rgba(255,255,255,0.7); border:1px solid rgba(255,255,255,0.1); }
.btn-ghost:hover { background:rgba(255,255,255,0.1); color:#fff; }
.input { width:100%; padding:0.75rem 1rem; border-radius:12px; background:rgba(255,255,255,0.05); border:1px solid rgba(255,255,255,0.1); color:#fff; font-size:0.875rem; }
.input:focus { outline:none; border-color:#6366f1; box-shadow:0 0 0 3px rgba(99,102,241,0.3); }
.input::placeholder { color:rgba(255,255,255,0.3); }
.alert { padding:0.75rem 1rem; border-radius:10px; font-size:0.85rem; margin-bottom:1rem; }
.alert-error { background:rgba(239,68,68,0.15); color:#f87171; border:1px solid rgba(239,68,68,0.3); }
.alert-success { background:rgba(16,185,129,0.15); color:#34d399; border:1px solid rgba(16,185,129,0.3); }
.badge { display:inline-flex; align-items:center; gap:0.35rem; padding:0.3rem 0.75rem; border-radius:9999px; font-size:0.75rem; font-weight:500; }
.badge-running { background:rgba(16,185,129,0.15); color:#34d399; border:1px solid rgba(16,185,129,0.3); }
.badge-stopped { background:rgba(239,68,68,0.15); color:#f87171; border:1px solid rgba(239,68,68,0.3); }
.check-label { display:flex; align-items:center; gap:0.5rem; padding:0.75rem 1rem; border-radius:12px; background:rgba(255,255,255,0.03); border:1px solid rgba(255,255,255,0.08); cursor:pointer; transition:all 0.2s; }
.check-label:has(input:checked) { background:rgba(16,185,129,0.1); border-color:rgba(16,185,129,0.3); }
.check-label input { display:none; }
.check-box { width:20px; height:20px; border-radius:6px; border:2px solid rgba(255,255,255,0.2); display:flex; align-items:center; justify-content:center; transition:all 0.2s; }
.check-label:has(input:checked) .check-box { background:#10b981; border-color:#10b981; }
</style>
</head>
<body>
<div class="bg-layer"></div>
<div class="container">
<div style="display:flex;align-items:center;gap:1rem;margin-bottom:2rem;">
  <a href="/" class="btn btn-ghost"><i class="fa-solid fa-arrow-left"></i>返回</a>
  <h1 style="font-size:1.5rem;font-weight:700;color:#fff;margin-left:auto;"><i class="fa-solid fa-pen-to-square" style="color:#6366f1;margin-right:0.5rem;"></i>编辑应用</h1>
</div>

{{if eq .Err "running"}}
<div class="alert alert-error"><i class="fa-solid fa-triangle-exclamation" style="margin-right:0.5rem;"></i>请先停止应用后再编辑</div>
{{end}}
{{if eq .Err "exists"}}
<div class="alert alert-error"><i class="fa-solid fa-triangle-exclamation" style="margin-right:0.5rem;"></i>新名称已存在</div>
{{end}}
{{if eq .Err "rename"}}
<div class="alert alert-error"><i class="fa-solid fa-triangle-exclamation" style="margin-right:0.5rem;"></i>重命名失败，请检查权限</div>
{{end}}
{{if eq .Err "save"}}
<div class="alert alert-error"><i class="fa-solid fa-triangle-exclamation" style="margin-right:0.5rem;"></i>保存配置失败</div>
{{end}}
{{if eq .Msg "1"}}
<div class="alert alert-success"><i class="fa-solid fa-check" style="margin-right:0.5rem;"></i>已保存</div>
{{end}}

<div class="glass" style="padding:1.5rem;">
  <div style="display:flex;align-items:center;justify-content:space-between;margin-bottom:1.5rem;">
    <h3 style="font-size:1.1rem;font-weight:600;color:#fff;">{{.Name}}</h3>
    {{if eq .Status "运行中"}}
    <span class="badge badge-running"><span style="width:6px;height:6px;background:#34d399;border-radius:50%;"></span>运行中</span>
    {{else}}
    <span class="badge badge-stopped"><span style="width:6px;height:6px;background:#f87171;border-radius:50%;"></span>已停止</span>
    {{end}}
  </div>

  <form id="editForm" action="/edit/{{.Name}}" method="post" style="display:grid;gap:0.75rem;">
    <div>
      <label style="font-size:0.75rem;color:rgba(255,255,255,0.5);margin-bottom:0.35rem;display:block;">应用名称（修改将重命名沙盒目录）</label>
      <input name="name" value="{{.Name}}" required class="input">
    </div>
    <div>
      <label style="font-size:0.75rem;color:rgba(255,255,255,0.5);margin-bottom:0.35rem;display:block;">沙盒路径</label>
      <input name="path" value="{{.Path}}" placeholder="留空不修改" class="input">
    </div>
    <div>
      <label style="font-size:0.75rem;color:rgba(255,255,255,0.5);margin-bottom:0.35rem;display:block;">启动命令</label>
      <input name="cmd" value="{{.Cmd}}" required class="input">
    </div>
    <label class="check-label">
      <input type="checkbox" name="auto"{{if .Auto}} checked{{end}}>
      <div class="check-box"><i class="fa-solid fa-check" style="font-size:0.7rem;color:#fff;"></i></div>
      <span style="color:rgba(255,255,255,0.7);font-size:0.875rem;">开机自启（面板重启后自动启动此应用，崩溃自动重启）</span>
    </label>
    <div style="display:grid;grid-template-columns:1fr 1fr;gap:0.75rem;margin-top:0.5rem;">
      <button class="btn btn-primary" style="width:100%;"><i class="fa-solid fa-save"></i>保存</button>
      <a href="/" class="btn btn-ghost" style="width:100%;"><i class="fa-solid fa-xmark"></i>取消</a>
    </div>
  </form>
</div>

<div style="font-size:0.75rem;color:rgba(255,255,255,0.3);margin-top:1rem;text-align:center;">
  <i class="fa-solid fa-clock" style="margin-right:0.3rem;"></i>创建时间：{{.Created}}
</div>
</div>
<script>
var ef = document.getElementById('editForm');
if (ef) {
  ef.addEventListener('submit', function(e) {
    e.preventDefault();
    var btn = ef.querySelector('button[type="submit"]');
    if (btn && !btn.disabled) { btn.disabled = true; btn.dataset.orig = btn.innerHTML; btn.innerHTML = '<i class="fa-solid fa-spinner fa-spin"></i> 保存中...'; }
    var fd = new FormData(ef);
    fetch(ef.action, { method: 'POST', body: fd })
      .then(function(r) { if (r.ok) { window.location.reload(); } else { alert('保存失败'); btn.disabled = false; btn.innerHTML = btn.dataset.orig; } })
      .catch(function() { alert('网络错误'); btn.disabled = false; btn.innerHTML = btn.dataset.orig; });
  });
}
</script>
</body>
</html>`

const htmlStore = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width,initial-scale=1.0">
<title>应用商店 - LightPanel</title>
<script src="https://cdn.tailwindcss.com"></script>
<link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.5.1/css/all.min.css">
<style>
@import url('https://fonts.googleapis.com/css2?family=Inter:wght@300;400;500;600;700&display=swap');
* { margin:0; padding:0; box-sizing:border-box; font-family:'Inter',system-ui,sans-serif; }
body { background:#0f0f23; min-height:100vh; position:relative; }
.bg-layer { position:fixed; inset:0; z-index:0; background:radial-gradient(ellipse at 50% 30%,rgba(168,85,247,0.15) 0%,transparent 60%),#0f0f23; }
.container { position:relative; z-index:1; max-width:900px; margin:0 auto; padding:1.5rem; }
.glass-card { background:rgba(255,255,255,0.03); backdrop-filter:blur(16px); border:1px solid rgba(255,255,255,0.08); border-radius:16px; padding:1.25rem; transition:all 0.3s ease; }
.glass-card:hover { background:rgba(255,255,255,0.06); border-color:rgba(255,255,255,0.15); transform:translateY(-2px); }
.btn { display:inline-flex; align-items:center; justify-content:center; gap:0.5rem; padding:0.6rem 1.2rem; border-radius:12px; font-size:0.875rem; font-weight:500; transition:all 0.2s ease; cursor:pointer; border:none; text-decoration:none; }
.btn-primary { background:linear-gradient(135deg,#6366f1,#7c3aed); color:#fff; }
.btn-primary:hover { box-shadow:0 4px 20px rgba(99,102,241,0.4); }
.btn-success { background:linear-gradient(135deg,#10b981,#059669); color:#fff; }
.btn-ghost { background:rgba(255,255,255,0.05); color:rgba(255,255,255,0.7); border:1px solid rgba(255,255,255,0.1); }
.btn-ghost:hover { background:rgba(255,255,255,0.1); color:#fff; }
.btn:disabled { opacity:0.5; cursor:not-allowed; }
.source-tabs { display:flex; gap:0.5rem; margin-bottom:1.5rem; flex-wrap:wrap; }
.source-tab { padding:0.5rem 1rem; border-radius:10px; font-size:0.8rem; transition:all 0.2s; text-decoration:none; }
.source-tab.active { background:linear-gradient(135deg,#6366f1,#7c3aed); color:#fff; }
.source-tab:not(.active) { background:rgba(255,255,255,0.05); color:rgba(255,255,255,0.5); border:1px solid rgba(255,255,255,0.08); }
.store-item { display:flex; align-items:center; gap:1rem; }
.store-icon { width:56px; height:56px; border-radius:14px; object-fit:cover; border:2px solid rgba(255,255,255,0.1); }
.overlay { position:fixed; inset:0; z-index:9998; background:rgba(0,0,0,0.6); backdrop-filter:blur(4px); display:flex; align-items:center; justify-content:center; }
.overlay-box { background:#1a1a2e; border:1px solid rgba(255,255,255,0.1); border-radius:20px; padding:2rem; max-width:400px; width:90%; text-align:center; }
@keyframes spin { to { transform:rotate(360deg); } }
.spinner { width:32px; height:32px; border:3px solid rgba(255,255,255,0.1); border-top-color:#10b981; border-radius:50%; animation:spin 0.8s linear infinite; margin:0 auto 1rem; }
@keyframes toast-in { from { transform:translateX(100%); opacity:0; } to { transform:translateX(0); opacity:1; } }
@keyframes toast-out { from { transform:translateX(0); opacity:1; } to { transform:translateX(100%); opacity:0; } }
.toast { position:fixed; top:1.5rem; right:1.5rem; z-index:9999; padding:0.85rem 1.25rem; border-radius:12px; font-size:0.85rem; font-weight:500; color:#fff; max-width:380px; box-shadow:0 8px 32px rgba(0,0,0,0.5); animation:toast-in 0.3s ease; }
.toast-success { background:linear-gradient(135deg,#10b981,#059669); }
.toast-error { background:linear-gradient(135deg,#ef4444,#dc2626); }
</style>
</head>
<body>
<div class="bg-layer"></div>
<div class="container">
<div style="display:flex;align-items:center;gap:1rem;margin-bottom:2rem;">
  <a href="/" class="btn btn-ghost"><i class="fa-solid fa-arrow-left"></i>返回</a>
  <a href="/source" class="btn btn-ghost"><i class="fa-solid fa-database"></i>源管理</a>
  <h1 style="font-size:1.5rem;font-weight:700;color:#fff;margin-left:auto;"><i class="fa-solid fa-store" style="color:#a855f7;margin-right:0.5rem;"></i>应用商店</h1>
</div>
<div class="source-tabs">
{{range $i,$s := .Sources}}
{{if eq $i $.Active}}
<span class="source-tab active">{{$s.Name}}</span>
{{else}}
<a href="/store?source={{$i}}" class="source-tab">{{$s.Name}}</a>
{{end}}
{{end}}
</div>
<div style="display:grid;gap:1rem;">
{{range $i,$a := .Apps}}
<div class="glass-card store-item">
  <img src="{{$a.Icon}}" class="store-icon" referrerpolicy="no-referrer" onerror="this.style.display='none'">
  <div style="flex:1;">
    <h3 style="font-size:1.05rem;font-weight:600;color:#fff;">{{$a.Name}}</h3>
    <p style="font-size:0.8rem;color:rgba(255,255,255,0.5);margin-top:0.25rem;">{{$a.Desc}}</p>
    <p style="font-size:0.7rem;color:rgba(255,255,255,0.3);margin-top:0.35rem;"><i class="fa-solid fa-user" style="margin-right:0.3rem;"></i>{{$a.Author}}</p>
  </div>
  <form action="/install/{{$i}}?source={{$.Active}}" method="post" style="display:inline;" class="install-form">
    <button class="btn btn-success"><i class="fa-solid fa-download"></i>部署</button>
  </form>
</div>
{{end}}
</div>
</div>

<script>
document.querySelectorAll('.install-form').forEach(function(form) {
  form.addEventListener('submit', function(e) {
    var btn = form.querySelector('button');
    if (btn && btn.disabled) { e.preventDefault(); return; }
  });
});
</script>
</body>
</html>`

const htmlDownloads = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width,initial-scale=1.0">
<title>下载管理 - LightPanel</title>
<script src="https://cdn.tailwindcss.com"></script>
<link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.5.1/css/all.min.css">
<style>
@import url('https://fonts.googleapis.com/css2?family=Inter:wght@300;400;500;600;700&display=swap');
* { margin:0; padding:0; box-sizing:border-box; font-family:'Inter',system-ui,sans-serif; }
body { background:#0f0f23; min-height:100vh; position:relative; }
.bg-layer { position:fixed; inset:0; z-index:0; background:radial-gradient(ellipse at 50% 50%,rgba(99,102,241,0.1) 0%,transparent 60%),#0f0f23; }
.container { position:relative; z-index:1; max-width:900px; margin:0 auto; padding:1.5rem; }
.glass { background:rgba(255,255,255,0.03); backdrop-filter:blur(16px); border:1px solid rgba(255,255,255,0.08); border-radius:16px; }
.glass-card { background:rgba(255,255,255,0.03); backdrop-filter:blur(16px); border:1px solid rgba(255,255,255,0.08); border-radius:16px; padding:1.25rem; margin-bottom:0.75rem; }
.btn { display:inline-flex; align-items:center; justify-content:center; gap:0.5rem; padding:0.5rem 1rem; border-radius:10px; font-size:0.8rem; font-weight:500; transition:all 0.2s ease; cursor:pointer; border:none; text-decoration:none; }
.btn-ghost { background:rgba(255,255,255,0.05); color:rgba(255,255,255,0.7); border:1px solid rgba(255,255,255,0.1); }
.btn-ghost:hover { background:rgba(255,255,255,0.1); color:#fff; }
.btn-success { background:linear-gradient(135deg,#10b981,#059669); color:#fff; }
.btn-warning { background:linear-gradient(135deg,#f59e0b,#d97706); color:#fff; }
.btn-danger { background:linear-gradient(135deg,#ef4444,#dc2626); color:#fff; }
.btn-info { background:linear-gradient(135deg,#3b82f6,#2563eb); color:#fff; }
.btn:disabled { opacity:0.4; cursor:not-allowed; }
.badge { display:inline-block; padding:0.2rem 0.6rem; border-radius:6px; font-size:0.7rem; font-weight:500; }
.badge-downloading { background:rgba(59,130,246,0.15); color:#60a5fa; border:1px solid rgba(59,130,246,0.3); }
.badge-paused { background:rgba(245,158,11,0.15); color:#fbbf24; border:1px solid rgba(245,158,11,0.3); }
.badge-completed { background:rgba(16,185,129,0.15); color:#34d399; border:1px solid rgba(16,185,129,0.3); }
.badge-error { background:rgba(239,68,68,0.15); color:#f87171; border:1px solid rgba(239,68,68,0.3); }
.progress-bar { width:100%; height:8px; background:rgba(255,255,255,0.08); border-radius:4px; overflow:hidden; margin:0.5rem 0; }
.progress-fill { height:100%; background:linear-gradient(90deg,#6366f1,#a855f7); border-radius:4px; transition:width 0.5s ease; }
</style>
</head>
<body>
<div class="bg-layer"></div>
<div class="container">
<div style="display:flex;align-items:center;gap:1rem;margin-bottom:2rem;">
  <a href="/" class="btn btn-ghost"><i class="fa-solid fa-arrow-left"></i>返回</a>
  <h1 style="font-size:1.5rem;font-weight:700;color:#fff;margin-left:auto;"><i class="fa-solid fa-download" style="color:#6366f1;margin-right:0.5rem;"></i>下载管理</h1>
</div>

<div id="taskList">
{{range .}}
<div class="glass-card" id="task-{{.ID}}">
  <div style="display:flex;justify-content:space-between;align-items:center;margin-bottom:0.5rem;">
    <div>
      <span style="font-size:1rem;font-weight:600;color:#fff;">{{.Name}}</span>
      <span class="badge badge-{{.Status}}" style="margin-left:0.5rem;">
        {{if eq .Status "downloading"}}下载中{{end}}
        {{if eq .Status "paused"}}已暂停{{end}}
        {{if eq .Status "completed"}}已完成{{end}}
        {{if eq .Status "error"}}失败{{end}}
      </span>
    </div>
    <div style="display:flex;gap:0.4rem;">
      {{if eq .Status "downloading"}}
      <button class="btn btn-warning" onclick="dlAction('{{.ID}}','pause')"><i class="fa-solid fa-pause"></i>暂停</button>
      {{end}}
      {{if eq .Status "paused"}}
      <button class="btn btn-success" onclick="dlAction('{{.ID}}','resume')"><i class="fa-solid fa-play"></i>继续</button>
      {{end}}
      {{if eq .Status "completed"}}
      <button class="btn btn-info" onclick="dlAction('{{.ID}}','install')"><i class="fa-solid fa-rocket"></i>安装</button>
      {{end}}
      <button class="btn btn-danger" onclick="dlAction('{{.ID}}','delete')"><i class="fa-solid fa-trash"></i></button>
    </div>
  </div>
  <div style="font-size:0.75rem;color:rgba(255,255,255,0.4);word-break:break-all;margin-bottom:0.3rem;">{{.URL}}</div>
  <div class="progress-bar"><div class="progress-fill" style="width:{{.Progress}}%"></div></div>
  <div style="font-size:0.75rem;color:rgba(255,255,255,0.4);display:flex;justify-content:space-between;">
    <span>{{.Progress}}%</span>
    <span>{{formatSize .Downloaded}} / {{formatSize .Size}}</span>
  </div>
</div>
{{end}}
{{if eq (len .) 0}}
<div class="glass" style="padding:3rem;text-align:center;">
  <i class="fa-solid fa-inbox" style="font-size:3rem;color:rgba(255,255,255,0.15);margin-bottom:1rem;"></i>
  <p style="color:rgba(255,255,255,0.4);">暂无下载任务</p>
</div>
{{end}}
</div>
</div>

<script>
function formatSize(b) {
  if (b <= 0) return '0 B';
  var u = ['B','KB','MB','GB'];
  var i = 0;
  while (b >= 1024 && i < u.length-1) { b /= 1024; i++; }
  return b.toFixed(1) + ' ' + u[i];
}

function dlAction(id, act) {
  fetch('/dl/action/' + id + '/' + act, { method: 'POST' })
    .then(function() { location.reload(); });
}

setInterval(function() {
  fetch('/dl/api')
    .then(function(r) { return r.json(); })
    .then(function(tasks) {
      tasks.forEach(function(t) {
        var card = document.getElementById('task-' + t.id);
        if (!card) return;
        var fill = card.querySelector('.progress-fill');
        if (fill) fill.style.width = t.progress + '%';
        var info = card.querySelectorAll('span');
        if (info.length >= 2) {
          info[info.length - 2].textContent = t.progress + '%';
        }
      });
    });
}, 1000);
</script>
</body>
</html>`

const htmlSource = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width,initial-scale=1.0">
<title>源管理 - LightPanel</title>
<script src="https://cdn.tailwindcss.com"></script>
<link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.5.1/css/all.min.css">
<style>
@import url('https://fonts.googleapis.com/css2?family=Inter:wght@300;400;500;600;700&display=swap');
* { margin:0; padding:0; box-sizing:border-box; font-family:'Inter',system-ui,sans-serif; }
body { background:#0f0f23; min-height:100vh; position:relative; }
.bg-layer { position:fixed; inset:0; z-index:0; background:radial-gradient(ellipse at 70% 60%,rgba(16,185,129,0.12) 0%,transparent 60%),#0f0f23; }
.container { position:relative; z-index:1; max-width:700px; margin:0 auto; padding:1.5rem; }
.glass { background:rgba(255,255,255,0.03); backdrop-filter:blur(16px); border:1px solid rgba(255,255,255,0.08); border-radius:16px; }
.glass-card { background:rgba(255,255,255,0.03); backdrop-filter:blur(16px); border:1px solid rgba(255,255,255,0.08); border-radius:16px; padding:1.25rem; display:flex; justify-content:space-between; align-items:center; transition:all 0.3s ease; }
.glass-card:hover { background:rgba(255,255,255,0.06); }
.btn { display:inline-flex; align-items:center; justify-content:center; gap:0.5rem; padding:0.6rem 1.2rem; border-radius:12px; font-size:0.875rem; font-weight:500; transition:all 0.2s ease; cursor:pointer; border:none; text-decoration:none; }
.btn-primary { background:linear-gradient(135deg,#10b981,#059669); color:#fff; }
.btn-danger { background:linear-gradient(135deg,#ef4444,#dc2626); color:#fff; }
.btn-ghost { background:rgba(255,255,255,0.05); color:rgba(255,255,255,0.7); border:1px solid rgba(255,255,255,0.1); }
.input { width:100%; padding:0.75rem 1rem; border-radius:12px; background:rgba(255,255,255,0.05); border:1px solid rgba(255,255,255,0.1); color:#fff; font-size:0.875rem; }
.input:focus { outline:none; border-color:#10b981; box-shadow:0 0 0 3px rgba(16,185,129,0.3); }
.input::placeholder { color:rgba(255,255,255,0.3); }
</style>
</head>
<body>
<div class="bg-layer"></div>
<div class="container">
<div style="display:flex;align-items:center;gap:1rem;margin-bottom:2rem;">
  <a href="/" class="btn btn-ghost"><i class="fa-solid fa-arrow-left"></i>返回</a>
  <h1 style="font-size:1.5rem;font-weight:700;color:#fff;margin-left:auto;"><i class="fa-solid fa-database" style="color:#10b981;margin-right:0.5rem;"></i>商店源管理</h1>
</div>
<div class="glass" style="padding:1.5rem;margin-bottom:1.5rem;">
  <h3 style="font-size:1rem;font-weight:600;color:#fff;margin-bottom:1rem;"><i class="fa-solid fa-plus" style="color:#10b981;margin-right:0.5rem;"></i>添加源</h3>
  <form action="/source/add" method="post" style="display:grid;grid-template-columns:1fr 2fr auto;gap:0.75rem;align-items:end;">
    <input name="name" placeholder="源名称" required class="input">
    <input name="url" placeholder="JSON 地址" required class="input">
    <button class="btn btn-primary" style="height:44px;"><i class="fa-solid fa-plus"></i>添加</button>
  </form>
</div>
<div style="display:grid;gap:0.75rem;">
{{range $i,$s := .}}
<div class="glass-card">
  <div>
    <p style="font-weight:600;color:#fff;">{{$s.Name}}</p>
    <p style="font-size:0.75rem;color:rgba(255,255,255,0.4);margin-top:0.25rem;word-break:break-all;">{{$s.URL}}</p>
  </div>
  <form action="/source/del/{{$i}}" method="post" style="display:inline;"><button class="btn btn-danger" onclick="return confirm('删除此源？')"><i class="fa-solid fa-trash"></i></button></form>
</div>
{{end}}
</div>
</div>
</body>
</html>`

const htmlSetting = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width,initial-scale=1.0">
<title>设置 - LightPanel</title>
<script src="https://cdn.tailwindcss.com"></script>
<link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.5.1/css/all.min.css">
<style>
@import url('https://fonts.googleapis.com/css2?family=Inter:wght@300;400;500;600;700&display=swap');
* { margin:0; padding:0; box-sizing:border-box; font-family:'Inter',system-ui,sans-serif; }
body { background:#0f0f23; min-height:100vh; position:relative; }
.bg-layer { position:fixed; inset:0; z-index:0; background:radial-gradient(ellipse at 30% 70%,rgba(245,158,11,0.12) 0%,transparent 60%),#0f0f23; }
.container { position:relative; z-index:1; max-width:550px; margin:0 auto; padding:1.5rem; }
.glass { background:rgba(255,255,255,0.03); backdrop-filter:blur(16px); border:1px solid rgba(255,255,255,0.08); border-radius:16px; }
.btn { display:inline-flex; align-items:center; justify-content:center; gap:0.5rem; padding:0.6rem 1.2rem; border-radius:12px; font-size:0.875rem; font-weight:500; transition:all 0.2s ease; cursor:pointer; border:none; text-decoration:none; }
.btn-primary { background:linear-gradient(135deg,#f59e0b,#d97706); color:#fff; }
.btn-ghost { background:rgba(255,255,255,0.05); color:rgba(255,255,255,0.7); border:1px solid rgba(255,255,255,0.1); }
.input { width:100%; padding:0.75rem 1rem; border-radius:12px; background:rgba(255,255,255,0.05); border:1px solid rgba(255,255,255,0.1); color:#fff; font-size:0.875rem; }
.input:focus { outline:none; border-color:#f59e0b; box-shadow:0 0 0 3px rgba(245,158,11,0.3); }
.input::placeholder { color:rgba(255,255,255,0.3); }
.alert { padding:0.75rem 1rem; border-radius:10px; font-size:0.85rem; margin-bottom:1rem; }
.alert-success { background:rgba(16,185,129,0.15); color:#34d399; border:1px solid rgba(16,185,129,0.3); }
.alert-error { background:rgba(239,68,68,0.15); color:#f87171; border:1px solid rgba(239,68,68,0.3); }
.radio-group { display:flex; gap:0.75rem; margin-bottom:0.75rem; }
.radio-label { display:flex; align-items:center; gap:0.4rem; padding:0.5rem 1rem; border-radius:10px; cursor:pointer; font-size:0.85rem; transition:all 0.2s; border:1px solid rgba(255,255,255,0.1); }
.radio-label:has(input:checked) { background:rgba(245,158,11,0.15); border-color:rgba(245,158,11,0.4); color:#fbbf24; }
.radio-label input { display:none; }
</style>
</head>
<body>
<div class="bg-layer"></div>
<div class="container">
<div style="display:flex;align-items:center;gap:1rem;margin-bottom:2rem;">
  <a href="/" class="btn btn-ghost"><i class="fa-solid fa-arrow-left"></i>返回</a>
  <h1 style="font-size:1.5rem;font-weight:700;color:#fff;margin-left:auto;"><i class="fa-solid fa-gear" style="color:#f59e0b;margin-right:0.5rem;"></i>设置</h1>
</div>

{{if eq .Msg "1"}}
<div class="alert alert-success"><i class="fa-solid fa-check" style="margin-right:0.5rem;"></i>设置已保存</div>
{{end}}
{{if eq .Err "password"}}
<div class="alert alert-error"><i class="fa-solid fa-xmark" style="margin-right:0.5rem;"></i>旧密码错误</div>
{{end}}

<div class="glass" style="padding:1.5rem;margin-bottom:1.5rem;">
  <h3 style="font-size:1rem;font-weight:600;color:#fff;margin-bottom:1rem;"><i class="fa-solid fa-key" style="color:#f59e0b;margin-right:0.5rem;"></i>修改密码</h3>
  <form action="/setting/save" method="post" style="display:grid;gap:0.75rem;">
    <input type="password" name="old" placeholder="当前密码" class="input">
    <input type="password" name="new" placeholder="新密码（留空不修改）" class="input">
    <button class="btn btn-primary" style="width:100%;"><i class="fa-solid fa-save"></i>保存密码</button>
  </form>
</div>

<div class="glass" style="padding:1.5rem;">
  <h3 style="font-size:1rem;font-weight:600;color:#fff;margin-bottom:1rem;"><i class="fa-solid fa-image" style="color:#f59e0b;margin-right:0.5rem;"></i>背景设置</h3>
  <form action="/setting/save" method="post" style="display:grid;gap:0.75rem;">
    <div class="radio-group">
      <label class="radio-label"><input type="radio" name="bg_type" value="gradient"{{if eq .BgType "gradient"}} checked{{end}}>渐变背景</label>
      <label class="radio-label"><input type="radio" name="bg_type" value="image"{{if eq .BgType "image"}} checked{{end}}>图片背景</label>
    </div>
    <input type="url" name="bg_url" placeholder="图片 API 地址，如 https://api.btstu.cn/sjbz/api.php" value="{{.BgUrl}}" class="input">
    <button class="btn btn-primary" style="width:100%;"><i class="fa-solid fa-save"></i>保存背景</button>
  </form>
</div>
</div>
</body>
</html>`

const htmlSystem = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="UTF-8">
<meta http-equiv="refresh" content="5">
<meta name="viewport" content="width=device-width,initial-scale=1.0">
<title>系统监控 - LightPanel</title>
<script src="https://cdn.tailwindcss.com"></script>
<link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.5.1/css/all.min.css">
<style>
@import url('https://fonts.googleapis.com/css2?family=Inter:wght@300;400;500;600;700&family=JetBrains+Mono:wght@400;500&display=swap');
* { margin:0; padding:0; box-sizing:border-box; font-family:'Inter',system-ui,sans-serif; }
body { background:#0f0f23; min-height:100vh; position:relative; }
.bg-layer { position:fixed; inset:0; z-index:0; background:radial-gradient(ellipse at 50% 50%,rgba(59,130,246,0.1) 0%,transparent 60%),#0f0f23; }
.container { position:relative; z-index:1; max-width:1000px; margin:0 auto; padding:1.5rem; }
.glass { background:rgba(255,255,255,0.03); backdrop-filter:blur(16px); border:1px solid rgba(255,255,255,0.08); border-radius:16px; overflow:hidden; }
.btn { display:inline-flex; align-items:center; justify-content:center; gap:0.5rem; padding:0.6rem 1.2rem; border-radius:12px; font-size:0.875rem; font-weight:500; transition:all 0.2s ease; cursor:pointer; border:none; text-decoration:none; }
.btn-ghost { background:rgba(255,255,255,0.05); color:rgba(255,255,255,0.7); border:1px solid rgba(255,255,255,0.1); }
table { width:100%; border-collapse:collapse; }
th { padding:0.85rem 1rem; text-align:left; font-size:0.7rem; text-transform:uppercase; letter-spacing:1px; color:rgba(255,255,255,0.4); border-bottom:1px solid rgba(255,255,255,0.06); }
td { padding:0.7rem 1rem; font-size:0.85rem; color:rgba(255,255,255,0.7); border-bottom:1px solid rgba(255,255,255,0.03); font-family:'JetBrains Mono',monospace; }
tr:hover td { background:rgba(255,255,255,0.02); }
.pid { color:#60a5fa; }
.kill { color:#f87171; background:none; border:none; padding:0; cursor:pointer; opacity:0.6; transition:opacity 0.2s; }
.kill:hover { opacity:1; }
</style>
</head>
<body>
<div class="bg-layer"></div>
<div class="container">
<div style="display:flex;align-items:center;gap:1rem;margin-bottom:1.5rem;">
  <a href="/" class="btn btn-ghost"><i class="fa-solid fa-arrow-left"></i>返回</a>
  <h1 style="font-size:1.5rem;font-weight:700;color:#fff;margin-left:auto;"><i class="fa-solid fa-chart-line" style="color:#3b82f6;margin-right:0.5rem;"></i>系统进程 <span style="font-size:0.8rem;color:rgba(255,255,255,0.4);font-weight:400;">自动刷新 5s</span></h1>
</div>
<div class="glass" style="overflow-x:auto;">
<table>
<tr><th>PID</th><th>CPU%</th><th>MEM%</th><th>进程名</th><th>操作</th></tr>
{{range .}}
<tr>
<td class="pid">{{.PID}}</td>
<td>{{formatFloat .Cpu}}</td>
<td>{{formatFloat .Mem}}</td>
<td style="font-family:'Inter',sans-serif;">{{.Name}}</td>
<td><form action="/kill/{{.PID}}" method="post" style="display:inline;"><button class="kill" onclick="return confirm('结束进程 {{.PID}}？')" style="background:none;border:none;padding:0;cursor:pointer;"><i class="fa-solid fa-xmark"></i></button></form></td>
</tr>
{{end}}
</table>
</div>
</div>
</body>
</html>`

const htmlLog = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width,initial-scale=1.0">
<title>日志 - LightPanel</title>
<script src="https://cdn.tailwindcss.com"></script>
<link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.5.1/css/all.min.css">
<style>
@import url('https://fonts.googleapis.com/css2?family=Inter:wght@300;400;500;600;700&family=JetBrains+Mono:wght@400;500&display=swap');
* { margin:0; padding:0; box-sizing:border-box; font-family:'Inter',system-ui,sans-serif; }
body { background:#0f0f23; min-height:100vh; position:relative; }
.bg-layer { position:fixed; inset:0; z-index:0; background:radial-gradient(ellipse at 60% 40%,rgba(99,102,241,0.1) 0%,transparent 60%),#0f0f23; }
.container { position:relative; z-index:1; max-width:900px; margin:0 auto; padding:1.5rem; }
.btn { display:inline-flex; align-items:center; justify-content:center; gap:0.5rem; padding:0.6rem 1.2rem; border-radius:12px; font-size:0.875rem; font-weight:500; transition:all 0.2s ease; cursor:pointer; border:none; text-decoration:none; }
.btn-ghost { background:rgba(255,255,255,0.05); color:rgba(255,255,255,0.7); border:1px solid rgba(255,255,255,0.1); }
.btn-danger { background:linear-gradient(135deg,#ef4444,#dc2626); color:#fff; }
.btn-sm { padding:0.4rem 0.8rem; font-size:0.75rem; border-radius:8px; }
.btn-active { background:rgba(99,102,241,0.2); color:#a5b4fc; border-color:rgba(99,102,241,0.4); }
.log-box { background:rgba(0,0,0,0.4); border:1px solid rgba(255,255,255,0.06); border-radius:12px; padding:1rem; height:75vh; overflow-y:auto; font-family:'JetBrains Mono',monospace; font-size:0.78rem; line-height:1.7; color:rgba(255,255,255,0.7); white-space:pre-wrap; word-break:break-all; }
.log-box::-webkit-scrollbar { width:6px; }
.log-box::-webkit-scrollbar-track { background:transparent; }
.log-box::-webkit-scrollbar-thumb { background:rgba(255,255,255,0.1); border-radius:3px; }
.log-error { color:#f87171; }
.log-warn { color:#fbbf24; }
.log-info { color:#60a5fa; }
.filter-bar { display:flex; gap:0.5rem; margin-bottom:1rem; flex-wrap:wrap; align-items:center; }
</style>
</head>
<body>
<div class="bg-layer"></div>
<div class="container">
<div style="display:flex;align-items:center;gap:1rem;margin-bottom:1.5rem;">
  <a href="/" class="btn btn-ghost"><i class="fa-solid fa-arrow-left"></i>返回</a>
  <h1 style="font-size:1.2rem;font-weight:600;color:#fff;margin-left:auto;"><i class="fa-solid fa-file-lines" style="color:#6366f1;margin-right:0.5rem;"></i>{{.Name}} 运行日志</h1>
  <form action="/log/clear/{{.Name}}" method="post" style="display:inline;"><button class="btn btn-danger btn-sm" onclick="return confirm('清空日志？')"><i class="fa-solid fa-trash"></i>清空</button></form>
</div>

<div class="filter-bar">
  <button class="btn btn-ghost btn-sm log-filter btn-active" data-filter="all"><i class="fa-solid fa-list"></i> 全部</button>
  <button class="btn btn-ghost btn-sm log-filter" data-filter="error"><i class="fa-solid fa-circle-xmark" style="color:#f87171;"></i> 错误</button>
  <button class="btn btn-ghost btn-sm log-filter" data-filter="warn"><i class="fa-solid fa-triangle-exclamation" style="color:#fbbf24;"></i> 警告</button>
  <button class="btn btn-ghost btn-sm log-filter" data-filter="info"><i class="fa-solid fa-circle-info" style="color:#60a5fa;"></i> 信息</button>
  <button class="btn btn-ghost btn-sm log-filter" data-filter="crash"><i class="fa-solid fa-skull" style="color:#ef4444;"></i> 崩溃</button>
  <span style="margin-left:auto;font-size:0.75rem;color:rgba(255,255,255,0.3);" id="logCount">{{.LineCount}} 行</span>
</div>

<pre class="log-box" id="logBox">{{.Log}}</pre>
</div>

<script>
var logBox = document.getElementById('logBox');
logBox.scrollTop = logBox.scrollHeight;

var filters = {
  all: null,
  error: [/error/i, /fail/i, /fatal/i, /panic/i, /exception/i, /crash/i, /cannot/i, /not found/i, /no such/i, /refused/i, /denied/i, /missing/i],
  warn: [/warn/i, /deprecated/i, /timeout/i, /retry/i, /slow/i],
  info: [/info/i, /start/i, /stop/i, /listen/i, /ready/i, /connect/i],
  crash: [/panic/i, /fatal/i, /segfault/i, /core dump/i, /killed/i, /oom/i, /out of memory/i]
};

document.querySelectorAll('.log-filter').forEach(function(btn) {
  btn.addEventListener('click', function(e) {
    e.preventDefault();
    var f = this.dataset.filter;
    document.querySelectorAll('.log-filter').forEach(function(b) { b.classList.remove('btn-active'); });
    this.classList.add('btn-active');

    var lines = logBox.textContent.split('\n');
    var pat = filters[f];
    if (!pat) {
      logBox.textContent = lines.join('\n');
    } else {
      var matched = [];
      for (var i = 0; i < lines.length; i++) {
        for (var j = 0; j < pat.length; j++) {
          if (pat[j].test(lines[i])) { matched.push(lines[i]); break; }
        }
      }
      logBox.textContent = matched.join('\n');
    }
    document.getElementById('logCount').textContent = (f === 'all' ? lines.length : (logBox.textContent ? logBox.textContent.split('\n').length : 0)) + ' 行';
  });
});
</script>
</body>
</html>`
