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
</div>
</div>
</div>
` + layoutJS + `
var currentTab = "download";
function switchTab(tab) {
	currentTab = tab;
	var dl = document.getElementById("downloadPanel");
	var ml = document.getElementById("manualPanel");
	var td = document.getElementById("tabDownload");
	var tm = document.getElementById("tabManual");
	var cb = document.getElementById("createBtn");
	if (tab === "download") {
		if (dl) dl.style.display = "block";
		if (ml) ml.style.display = "none";
		if (td) td.className = "btn btn-primary btn-sm";
		if (tm) tm.className = "btn btn-ghost btn-sm";
		if (cb) cb.innerHTML = "<i class=\"fa-solid fa-plus\"></i>创建应用";
	} else {
		if (dl) dl.style.display = "none";
		if (ml) ml.style.display = "block";
		if (td) td.className = "btn btn-ghost btn-sm";
		if (tm) tm.className = "btn btn-primary btn-sm";
		if (cb) cb.innerHTML = "<i class=\"fa-solid fa-plus\"></i>添加应用";
	}
}
var form = document.getElementById("createForm");
if (form) {
	form.addEventListener("submit", function(e) {
		e.preventDefault();
		var isManual = currentTab === "manual";
		var nameEl = isManual ? document.getElementById("manualName") : document.getElementById("appName");
		var prog = document.getElementById("createProgress");
		var fill = document.getElementById("progressFill");
		var text = document.getElementById("progressText");
		if (!nameEl.value.trim()) { alert("请输入应用名称"); return; }
		prog.style.display = "block";
		fill.style.width = "10%";
		text.textContent = isManual ? "添加应用..." : "准备创建...";
		var fd = new FormData();
		if (isManual) {
			fd.append("name", document.getElementById("manualName").value.trim());
			fd.append("path", document.getElementById("manualPath").value.trim());
			fd.append("cmd", document.getElementById("manualCmd").value.trim());
			fd.append("workdir", document.getElementById("manualWorkDir").value.trim());
			fd.append("url", document.getElementById("manualUrl").value.trim());
			fd.append("auto", document.getElementById("manualAuto").checked ? "1" : "0");
			fetch("/create/manual", {method: "POST", body: fd}).then(function(r) { return r.json(); }).then(function(data) {
				fill.style.width = "100%";
				text.textContent = data.error || "完成";
				if (!data.error) { setTimeout(function() { location.reload(); }, 500); }
				else { prog.style.display = "none"; alert(data.error); }
			}).catch(function(e) { prog.style.display = "none"; alert("请求失败: " + e); });
		} else {
			fd.append("name", document.getElementById("appName").value.trim());
			fd.append("cmd", document.getElementById("appCmd").value.trim());
			fd.append("url", document.getElementById("appUrl").value.trim());
			fd.append("setup_cmd", document.getElementById("appSetup").value.trim());
			fd.append("auto_extract", document.getElementById("autoExtract").checked ? "on" : "");
			fd.append("make_exec", document.getElementById("makeExec").checked ? "on" : "");
			fetch("/create", {method: "POST", body: fd}).then(function(r) { return r.json(); }).then(function(data) {
				if (data.error) { fill.style.width = "0%"; prog.style.display = "none"; alert(data.error); return; }
				fill.style.width = "30%";
				text.textContent = "正在下载...";
				var taskId = data.task;
				var pollInt = setInterval(function() {
					fetch("/create/progress/" + taskId).then(function(r) { return r.json(); }).then(function(t) {
						if (t && t.status) {
							fill.style.width = (t.progress || 0) + "%";
							text.textContent = t.message || t.status;
							if (t.status === "completed" || t.status === "error") {
								clearInterval(pollInt);
								if (t.status === "completed") { location.reload(); }
								else { prog.style.display = "none"; alert(t.message); }
							}
						}
					}).catch(function() { clearInterval(pollInt); });
				}, 500);
			}).catch(function(e) { prog.style.display = "none"; alert("请求失败"); });
		}
	});
}
</script>
</body>
</html>`