package handlers

var htmlLogin = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width,initial-scale=1.0">
<title>登录 - 朱雀面板</title>
<style>
*{
margin:0;padding:0;box-sizing:border-box;font-family:'Inter',system-ui,sans-serif}
:root{
--bg:#0f0f23;--bg-grad:rgba(229,62,62,0.15),rgba(192,48,48,0.1);
--card:rgba(255,255,255,0.04);--card-b:rgba(255,255,255,0.08);
--text:#fff;--text2:rgba(255,255,255,0.4);--text3:rgba(255,255,255,0.5);
--input-bg:rgba(255,255,255,0.05);--input-b:rgba(255,255,255,0.1);--input-t:#fff;
--btn-bg:linear-gradient(135deg,#e53e3e,#c53030);--btn-h:0 6px 20px rgba(229,62,62,0.35);
--err-bg:rgba(239,68,68,0.1);--err-c:#f87171;--err-b:rgba(239,68,68,0.2);
--ft:rgba(255,255,255,0.25)}
[data-theme="light"]{
--bg:#e8eaed;--bg-grad:rgba(229,62,62,0.08),rgba(192,48,48,0.05);
--card:rgba(255,255,255,0.7);--card-b:rgba(0,0,0,0.1);
--text:#1a1a2e;--text2:rgba(0,0,0,0.55);--text3:rgba(0,0,0,0.6);
--input-bg:rgba(255,255,255,0.7);--input-b:rgba(0,0,0,0.12);--input-t:#1a1a2e;
--btn-bg:linear-gradient(135deg,#e53e3e,#c53030);--btn-h:0 6px 20px rgba(229,62,62,0.25);
--err-bg:rgba(239,68,68,0.08);--err-c:#dc2626;--err-b:rgba(239,68,68,0.15);
--ft:rgba(0,0,0,0.25)}
body{min-height:100vh;display:flex;align-items:center;justify-content:center;background:var(--bg);position:relative;overflow:hidden}
body::before{content:'';position:absolute;inset:0;background:radial-gradient(ellipse at 30% 50%,var(--bg-grad));opacity:0.9}
.card{position:relative;z-index:1;background:var(--card);backdrop-filter:blur(30px);border:1px solid var(--card-b);border-radius:20px;padding:2.5rem;width:360px}
.logo{width:48px;height:48px;background:linear-gradient(135deg,#e53e3e,#fc8181);border-radius:50%;display:flex;align-items:center;justify-content:center;margin:0 auto 1.2rem;box-shadow:0 0 15px rgba(229,62,62,0.4);border:2px solid rgba(255,255,255,0.2)}
[data-theme="light"] .logo{border:2px solid rgba(255,255,255,0.5)}
.logo i{color:#fff;font-size:1.3rem}
h1{text-align:center;font-size:1.3rem;font-weight:700;color:var(--text);margin-bottom:0.3rem}
.sub{text-align:center;font-size:0.8rem;color:var(--text2);margin-bottom:1.8rem}
.field{margin-bottom:0.8rem}
.field label{display:block;font-size:0.75rem;font-weight:600;color:var(--text3);margin-bottom:0.3rem}
.field input{width:100%;padding:0.65rem 0.85rem;border-radius:10px;border:1px solid var(--input-b);background:var(--input-bg);color:var(--input-t);font-size:0.85rem;outline:none;transition:all 0.2s}
.field input:focus{border-color:#e53e3e;box-shadow:0 0 0 3px rgba(229,62,62,0.15)}
.btn{width:100%;padding:0.75rem;border-radius:10px;border:none;background:var(--btn-bg);color:#fff;font-size:0.9rem;font-weight:600;cursor:pointer;transition:all 0.2s;margin-top:0.5rem}
.btn:hover{box-shadow:var(--btn-h)}
.btn:disabled{opacity:0.5;cursor:not-allowed}
.err{background:var(--err-bg);color:var(--err-c);padding:0.6rem 0.8rem;border-radius:8px;font-size:0.78rem;margin-bottom:0.8rem;text-align:center;border:1px solid var(--err-b)}
.ft{text-align:center;font-size:0.65rem;color:var(--ft);margin-top:1.2rem}
</style>
<link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.5.1/css/all.min.css">
</head>
<body>
<div class="card">
  <div class="logo"><i class="fa-solid fa-fire-flame-curved"></i></div>
  <h1>朱雀面板</h1>
  <p class="sub">服务器管理面板</p>
  {{if eq .Err "1"}}
  <div class="err"><i class="fa-solid fa-circle-exclamation" style="margin-right:0.3rem;"></i>用户名或密码错误</div>
  {{end}}
  {{if eq .Err "locked"}}
  <div class="err"><i class="fa-solid fa-lock" style="margin-right:0.3rem;"></i>登录失败次数过多，请等待 <span id="lockTimer">{{.LockTime}}</span> 秒后重试</div>
  {{end}}
  <form action="/login/auth" method="post" onsubmit="var b=document.getElementById('loginBtn');if(b&&b.disabled){event.preventDefault();return false}">
    <div class="field"><label>用户名</label><input name="username" placeholder="请输入用户名" required autocomplete="username"></div>
    <div class="field"><label>密码</label><input name="password" type="password" placeholder="请输入密码" required autocomplete="current-password"></div>
    <button class="btn" type="submit" id="loginBtn"><i class="fa-solid fa-right-to-bracket" style="margin-right:0.4rem;"></i>登录</button>
  </form>
  <div class="ft">朱雀面板 · 轻量高效</div>
</div>
{{if eq .Err "locked"}}
<script>
(function(){
var el=document.getElementById('lockTimer');
var btn=document.getElementById('loginBtn');
if(!el||!btn)return;
var s=parseInt(el.textContent)||0;
btn.disabled=true;
var t=setInterval(function(){
s--;
if(s<=0){clearInterval(t);btn.disabled=false;el.textContent='0';}
else{el.textContent=s;}
},1000);
})();
</script>
{{end}}
<script>
(function(){
var t=localStorage.getItem('lp_theme');
if(t==='light'){document.documentElement.setAttribute('data-theme','light');}
else if(t==='dark'){document.documentElement.setAttribute('data-theme','dark');}
else if(window.matchMedia('(prefers-color-scheme:light)').matches){document.documentElement.setAttribute('data-theme','light');}
})();
</script>
</body>
</html>`
