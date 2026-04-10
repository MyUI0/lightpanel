package handlers

const layoutCSS = `@import url('https://fonts.googleapis.com/css2?family=Inter:wght@300;400;500;600;700&display=swap');
:root{--accent:#e53e3e;--bg:#0f0f23;--bg2:rgba(15,15,35,0.35);--card:rgba(255,255,255,0.03);--card-h:rgba(255,255,255,0.06);--text:#e2e8f0;--text2:rgba(255,255,255,0.45);--border:rgba(255,255,255,0.06);--input-bg:rgba(255,255,255,0.04);--input-b:rgba(255,255,255,0.08);--input-t:#e2e8f0;--input-ph:rgba(255,255,255,0.3);--sidebar-w:200px;--sidebar-cw:70px}
[data-theme="dark"]{--bg:#0f0f23;--bg2:rgba(15,15,35,0.35);--card:rgba(255,255,255,0.03);--card-h:rgba(255,255,255,0.06);--text:#e2e8f0;--text2:rgba(255,255,255,0.45);--border:rgba(255,255,255,0.06);--input-bg:rgba(255,255,255,0.04);--input-b:rgba(255,255,255,0.08);--input-t:#e2e8f0;--input-ph:rgba(255,255,255,0.3)}
[data-theme="light"]{--bg:#e8eaed;--bg2:rgba(255,255,255,0.55);--card:rgba(255,255,255,0.6);--card-h:rgba(255,255,255,0.8);--text:#1a1a2e;--text2:rgba(0,0,0,0.55);--border:rgba(0,0,0,0.1);--input-bg:rgba(255,255,255,0.7);--input-b:rgba(0,0,0,0.12);--input-t:#1a1a2e;--input-ph:rgba(0,0,0,0.35)}
@media(prefers-color-scheme:light){:root:not([data-theme="dark"]){--bg:#e8eaed;--bg2:rgba(255,255,255,0.55);--card:rgba(255,255,255,0.6);--card-h:rgba(255,255,255,0.8);--text:#1a1a2e;--text2:rgba(0,0,0,0.55);--border:rgba(0,0,0,0.1);--input-bg:rgba(255,255,255,0.7);--input-b:rgba(0,0,0,0.12);--input-t:#1a1a2e;--input-ph:rgba(0,0,0,0.35)}}
*{margin:0;padding:0;box-sizing:border-box;font-family:'Inter',system-ui,sans-serif}
html,body{width:100%;height:100%;overflow:hidden}
body{background:var(--bg);min-height:100vh;color:var(--text);display:flex}
.bg-layer{position:fixed;inset:0;z-index:0;background-size:cover;background-position:center;background-repeat:no-repeat}
.bg-layer::before{content:'';position:absolute;inset:0;background:radial-gradient(ellipse at 50% 30%,rgba(229,62,62,0.12) 0%,transparent 60%),var(--bg);opacity:0.9}
.sidebar{width:var(--sidebar-w);min-height:100vh;background:var(--bg2);backdrop-filter:blur(30px);-webkit-backdrop-filter:blur(30px);border-right:1px solid var(--border);display:flex;flex-direction:column;position:fixed;left:0;top:0;bottom:0;z-index:100;transition:width 0.25s ease;overflow:hidden}
.sidebar.collapsed{width:var(--sidebar-cw)}
.sidebar .logo-row{display:flex;align-items:center;gap:0.6rem;padding:1.2rem 1rem;border-bottom:1px solid var(--border);justify-content:center}
.sidebar .logo-icon{width:36px;height:36px;border-radius:50%;display:flex;align-items:center;justify-content:center;flex-shrink:0;position:relative;overflow:hidden;border:2px solid rgba(229,62,62,0.4);box-shadow:0 0 15px rgba(229,62,62,0.3)}
.sidebar .logo-icon img{width:100%;height:100%;object-fit:cover;border-radius:50%}
.sidebar .logo-icon i{color:#fff;font-size:1rem}
.sidebar .logo-text{font-size:0.9rem;font-weight:700;color:var(--text);white-space:nowrap;transition:opacity 0.2s}
.sidebar.collapsed .logo-text{opacity:0;pointer-events:none;display:none}
.sidebar nav{flex:1;padding:0.5rem 0.4rem;display:flex;flex-direction:column;gap:2px}
.sidebar nav a{display:flex;align-items:center;gap:0.7rem;padding:0.65rem 0.7rem;border-radius:10px;color:var(--text2);text-decoration:none;font-size:0.8rem;font-weight:500;transition:all 0.15s;white-space:nowrap;justify-content:flex-start}
.sidebar nav a:hover{background:var(--card-h);color:var(--text)}
.sidebar nav a.active{background:rgba(229,62,62,0.15);color:#fc8181}
.sidebar nav a i{width:18px;text-align:center;font-size:0.9rem;flex-shrink:0}
.sidebar nav a .nav-long{transition:opacity 0.2s}
.sidebar nav a .nav-short{display:none;transition:opacity 0.2s}
.sidebar.collapsed nav{display:block;padding:0.5rem 0}
.sidebar.collapsed nav a{display:flex;justify-content:center;align-items:center;padding:0.35rem 0.3rem;gap:0;width:80%;margin:0 auto}
.sidebar.collapsed nav a i{display:none}
.sidebar.collapsed nav a .nav-long{display:none}
.sidebar.collapsed nav a .nav-short{display:block;font-size:0.82rem;font-weight:600;text-align:center;line-height:1}
.sidebar .side-footer{padding:0.6rem 0.4rem;border-top:1px solid var(--border);display:flex;flex-direction:column;gap:2px}
.sidebar .side-footer a{display:flex;align-items:center;gap:0.7rem;padding:0.6rem 0.7rem;border-radius:10px;color:var(--text2);text-decoration:none;font-size:0.8rem;transition:all 0.15s;white-space:nowrap;justify-content:flex-start}
.sidebar .side-footer a:hover{background:var(--card-h);color:var(--text)}
.sidebar .side-footer a i{width:18px;text-align:center;font-size:0.85rem;flex-shrink:0}
.sidebar .side-footer a .nav-long{transition:opacity 0.2s}
.sidebar .side-footer a .nav-short{display:none}
.sidebar.collapsed .side-footer{display:block}
.sidebar.collapsed .side-footer a{display:flex;justify-content:center;align-items:center;padding:0.35rem 0.3rem;gap:0;width:80%;margin:0 auto}
.sidebar.collapsed .side-footer a i{display:none}
.sidebar.collapsed .side-footer a .nav-long{display:none}
.sidebar.collapsed .side-footer a .nav-short{display:block;font-size:0.82rem;font-weight:600;text-align:center;line-height:1}
.sidebar .side-footer .logout{color:#f87171}
.main{margin-left:var(--sidebar-w);flex:1;min-height:100vh;display:flex;flex-direction:column;transition:margin-left 0.25s ease;position:relative;z-index:1}
.sidebar.collapsed~.main{margin-left:var(--sidebar-cw)}
.topbar{display:flex;align-items:center;gap:0.8rem;padding:0.7rem 1.5rem;border-bottom:1px solid var(--border);background:var(--bg2);backdrop-filter:blur(20px);-webkit-backdrop-filter:blur(20px);position:sticky;top:0;z-index:50}
.topbar .toggle-btn{width:30px;height:30px;border-radius:50%;border:1px solid var(--border);background:var(--card);color:var(--text2);display:flex;align-items:center;justify-content:center;cursor:pointer;transition:all 0.15s;flex-shrink:0;font-size:1.1rem;line-height:1}
.topbar .toggle-btn:hover{background:var(--card-h);color:var(--text)}
.topbar .page-title{font-size:0.95rem;font-weight:600;color:var(--text)}
.topbar .spacer{flex:1}
.content{flex:1;padding:1.2rem 1.5rem}
.card{background:var(--card);backdrop-filter:blur(30px);-webkit-backdrop-filter:blur(30px);border:1px solid var(--border);border-radius:14px;padding:1rem;transition:all 0.2s}
.card:hover{background:var(--card-h)}
.glass{background:var(--card);backdrop-filter:blur(30px);-webkit-backdrop-filter:blur(30px);border:1px solid var(--border);border-radius:14px}
.btn{display:inline-flex;align-items:center;justify-content:center;gap:0.4rem;padding:0.5rem 1rem;border-radius:9px;font-size:0.75rem;font-weight:500;transition:all 0.15s;cursor:pointer;border:none;text-decoration:none}
.btn-primary{background:linear-gradient(135deg,#e53e3e,#c53030);color:#fff}
.btn-primary:hover{box-shadow:0 3px 12px rgba(229,62,62,0.3)}
.btn-ghost{background:var(--card);color:var(--text2);border:1px solid var(--border)}
.btn-ghost:hover{background:var(--card-h);color:var(--text)}
.btn-danger{background:linear-gradient(135deg,#ef4444,#dc2626);color:#fff}
.btn-success{background:linear-gradient(135deg,#10b981,#059669);color:#fff}
.btn-warning{background:linear-gradient(135deg,#f59e0b,#d97706);color:#fff}
.file-input-btn{display:inline-flex;cursor:pointer}
.file-input-btn .file-name{cursor:pointer}

.btn-sm{padding:0.4rem 0.6rem;font-size:0.7rem;border-radius:6px;white-space:nowrap}
.input{width:100%;padding:0.6rem 0.8rem;border-radius:9px;background:var(--input-bg);border:1px solid var(--input-b);color:var(--input-t);font-size:0.8rem;outline:none;transition:all 0.2s}
.input:focus{border-color:#e53e3e;box-shadow:0 0 0 3px rgba(229,62,62,0.15)}
.input::placeholder{color:var(--input-ph)}
.alert{padding:0.6rem 0.8rem;border-radius:9px;font-size:0.78rem;margin-bottom:0.8rem}
.alert-error{background:rgba(239,68,68,0.1);color:#f87171;border:1px solid rgba(239,68,68,0.2)}
.alert-success{background:rgba(16,185,129,0.1);color:#34d399;border:1px solid rgba(16,185,129,0.2)}
.badge{display:inline-flex;align-items:center;gap:0.3rem;padding:0.2rem 0.6rem;border-radius:9999px;font-size:0.68rem;font-weight:500}
.badge-running{background:rgba(16,185,129,0.1);color:#34d399;border:1px solid rgba(16,185,129,0.2)}
.badge-stopped{background:rgba(239,68,68,0.1);color:#f87171;border:1px solid rgba(239,68,68,0.2)}
.stat-card{text-align:center;padding:0.8rem}
.stat-value{font-size:1.3rem;font-weight:700;color:var(--text)}
.stat-label{font-size:0.7rem;color:var(--text2);margin-top:0.2rem}
.progress-bar{width:100%;height:6px;background:rgba(255,255,255,0.06);border-radius:3px;overflow:hidden}
.progress-fill{height:100%;background:linear-gradient(90deg,#e53e3e,#fc8181);border-radius:3px;transition:width 0.3s}
.check-label{display:flex;align-items:center;gap:0.4rem;padding:0.6rem 0.8rem;border-radius:9px;background:var(--card);border:1px solid var(--border);cursor:pointer;transition:all 0.15s}
.check-label:has(input:checked){background:rgba(16,185,129,0.08);border-color:rgba(16,185,129,0.25)}
.check-label input{display:none}
.check-box{width:18px;height:18px;border-radius:5px;border:2px solid var(--input-b);display:flex;align-items:center;justify-content:center;transition:all 0.15s}
.check-label:has(input:checked) .check-box{background:#10b981;border-color:#10b981}
.check-label:has(input:checked) .check-box i{color:#fff;font-size:0.6rem}
.fail-card{background:rgba(239,68,68,0.08);border:1px solid rgba(239,68,68,0.2);border-radius:12px;padding:0.8rem;margin-bottom:0.8rem}
.fail-card h4{color:#f87171;font-size:0.82rem;margin-bottom:0.4rem}
.fail-card pre{font-size:0.7rem;color:var(--text2);white-space:pre-wrap;word-break:break-all;max-height:120px;overflow-y:auto}
.proc-wrap{overflow-x:auto}
.proc-table{width:100%;border-collapse:collapse}
.proc-table th,.proc-table td{padding:0.6rem 0.8rem;text-align:left;font-size:0.78rem;border-bottom:1px solid var(--border)}
.proc-table th{color:var(--text2);font-weight:500}
.proc-table td{color:var(--text)}`

const layoutJS = `<script>
function toggleEditForm(i){
var f=document.getElementById("edit-"+i);
if(f){f.style.display=f.style.display==="none"?"grid":"none";}
}
(function(){
var csrfCookie=document.cookie.match(/(?:^|; )lp_csrf=([^;]*)/);
var csrfToken=csrfCookie?decodeURIComponent(csrfCookie[1]):"";
if(csrfToken){
var forms=document.querySelectorAll('form[method="POST"],a[href]');
for(var i=0;i<forms.length;i++){
var f=forms[i];
var act=f.getAttribute("action")||"";
if(act.indexOf("/login/auth")===0||act.indexOf("/logout")===0)continue;
var inp=document.createElement("input");
inp.type="hidden";inp.name="csrf_token";inp.value=csrfToken;
f.appendChild(inp);
}
}
var sb=document.getElementById("sidebar");
var tb=document.getElementById("toggleBtn");
if(!sb||!tb)return;
var s=localStorage.getItem("lp_sidebar");
if(s==="collapsed"){sb.classList.add("collapsed");}
tb.addEventListener("click",function(){
sb.classList.toggle("collapsed");
localStorage.setItem("lp_sidebar",sb.classList.contains("collapsed")?"collapsed":"expanded");
});
var b=document.getElementById("themeBtn");
if(b){
var t=localStorage.getItem("lp_theme");
if(t==="dark"){applyTheme("dark");}
else if(t==="light"){applyTheme("light");}
b.addEventListener("click",function(){
var c=document.documentElement.getAttribute("data-theme");
if(c==="light"){applyTheme("dark");}
else{applyTheme("light");}
});
}
function applyTheme(mode){
if(mode==="light"){
document.documentElement.setAttribute("data-theme","light");
document.documentElement.style.setProperty("--bg","#e8eaed");
document.documentElement.style.setProperty("--bg2","rgba(255,255,255,0.55)");
document.documentElement.style.setProperty("--card","rgba(255,255,255,0.6)");
document.documentElement.style.setProperty("--card-h","rgba(255,255,255,0.8)");
document.documentElement.style.setProperty("--text","#1a1a2e");
document.documentElement.style.setProperty("--text2","rgba(0,0,0,0.55)");
document.documentElement.style.setProperty("--border","rgba(0,0,0,0.1)");
document.documentElement.style.setProperty("--input-bg","rgba(255,255,255,0.7)");
document.documentElement.style.setProperty("--input-b","rgba(0,0,0,0.12)");
document.documentElement.style.setProperty("--input-t","#1a1a2e");
document.documentElement.style.setProperty("--input-ph","rgba(0,0,0,0.35)");
var ic=b.querySelector("i");if(ic){ic.className="fa-solid fa-sun";}
localStorage.setItem("lp_theme","light");
}else{
document.documentElement.setAttribute("data-theme","dark");
document.documentElement.style.removeProperty("--bg");
document.documentElement.style.removeProperty("--bg2");
document.documentElement.style.removeProperty("--card");
document.documentElement.style.removeProperty("--card-h");
document.documentElement.style.removeProperty("--text");
document.documentElement.style.removeProperty("--text2");
document.documentElement.style.removeProperty("--border");
document.documentElement.style.removeProperty("--input-bg");
document.documentElement.style.removeProperty("--input-b");
document.documentElement.style.removeProperty("--input-t");
document.documentElement.style.removeProperty("--input-ph");
var ic=b.querySelector("i");if(ic){ic.className="fa-solid fa-moon";}
localStorage.setItem("lp_theme","dark");
}
}
})();
</script>`
