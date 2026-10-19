(() => {
  'use strict';
  const storage = { remove(key) { try { localStorage.removeItem(key); } catch {} }, get(key) { try { return localStorage.getItem(key); } catch { return null; } }, set(key,value) { try { localStorage.setItem(key,value); } catch {} } };
  window.LP = {
    csrf() { const match = document.cookie.match(/(?:^|; )lp_csrf=([^;]*)/); return match ? decodeURIComponent(match[1]) : ''; },
    async request(url, options = {}) {
      const headers = new Headers(options.headers);
      if (options.method && options.method.toUpperCase() !== 'GET') headers.set('X-CSRF-Token', LP.csrf());
      const response = await fetch(url, {...options, headers, signal: options.signal || AbortSignal.timeout(90000)});
      if (response.redirected && new URL(response.url).pathname === '/login') throw new Error('登录已过期，请重新登录');
      if (!response.ok) throw new Error((await response.text()).trim().slice(0,160) || '请求失败，请重试');
      return response;
    },
    confirm(message) {
      return new Promise(resolve => {
        const dialog=document.createElement('dialog'); dialog.className='drawer'; dialog.setAttribute('aria-label','确认操作');
        const heading=document.createElement('h2');heading.textContent='确认操作';heading.style.marginBottom='14px';
        const text=document.createElement('p');text.textContent=message;text.className='drawer-description';
        const actions=document.createElement('div');actions.className='heading-actions';
        const cancel=document.createElement('button');cancel.className='btn btn-ghost';cancel.textContent='取消';
        const accept=document.createElement('button');accept.className='btn btn-primary';accept.textContent='确认执行';
        const finish=value=>{dialog.close();dialog.remove();resolve(value);};
        cancel.addEventListener('click',()=>finish(false));accept.addEventListener('click',()=>finish(true));
        dialog.addEventListener('cancel',e=>{e.preventDefault();finish(false);});
        actions.append(cancel,accept);dialog.append(heading,text,actions);document.body.append(dialog);dialog.showModal();cancel.focus();
      });
    },
    toast(message, error = false) {
      document.getElementById('lpToast')?.remove();
      const toast = document.createElement('div');
      toast.id = 'lpToast'; toast.className = 'toast' + (error ? ' error' : '');
      toast.setAttribute('role', error ? 'alert' : 'status'); toast.textContent = message;
      document.body.append(toast); setTimeout(() => toast.remove(), error ? 9000 : 4500);
    }, storage
  };
  const savedTheme = storage.get('lp_theme');
  const systemTheme = window.matchMedia('(prefers-color-scheme: dark)');
  document.documentElement.dataset.theme = savedTheme || (systemTheme.matches ? 'dark' : 'light');
  systemTheme.addEventListener('change', e => { if (!storage.get('lp_theme')) document.documentElement.dataset.theme = e.matches ? 'dark' : 'light'; });
  window.toggleEditForm = i => { const f = document.getElementById('edit-'+i); if(f) f.style.display = f.style.display === 'none' ? 'grid' : 'none'; };
  document.addEventListener('DOMContentLoaded', () => {
    document.querySelectorAll('form').forEach(form => {
      if (form.method.toLowerCase() !== 'post' || form.action.includes('/login/auth')) return;
      let input = form.querySelector('[name=csrf_token]');
      if (!input) { input = document.createElement('input'); input.type='hidden'; input.name='csrf_token'; form.append(input); }
      input.value = LP.csrf();
    });
    document.querySelectorAll('input:not([type=hidden]),select,textarea').forEach(input => {
      if (!input.labels?.length && !input.hasAttribute('aria-label')) input.setAttribute('aria-label', input.placeholder || input.name || '输入');
    });
    const sidebar = document.getElementById('sidebar'), toggle = document.getElementById('toggleBtn');
    const narrow = window.matchMedia('(max-width:760px)');
    const shade = document.createElement('button'); shade.className='sidebar-shade'; shade.ariaLabel='关闭导航'; document.body.append(shade);
    const close = () => { sidebar?.classList.remove('mobile-open'); shade.classList.remove('visible'); toggle?.setAttribute('aria-expanded','false'); };
    const setNavigation = () => { close(); sidebar?.classList.toggle('collapsed', !narrow.matches && storage.get('lp_sidebar') === 'collapsed'); toggle?.setAttribute('aria-expanded',String(!narrow.matches && !sidebar?.classList.contains('collapsed'))); };
    setNavigation(); narrow.addEventListener('change',setNavigation);
    toggle?.addEventListener('click', () => {
      if (narrow.matches) { const open = sidebar.classList.toggle('mobile-open'); shade.classList.toggle('visible',open); toggle.setAttribute('aria-expanded',String(open)); }
      else { sidebar.classList.toggle('collapsed'); storage.set('lp_sidebar',sidebar.classList.contains('collapsed')?'collapsed':'expanded'); toggle.setAttribute('aria-expanded',String(!sidebar.classList.contains('collapsed'))); }
    });
    shade.addEventListener('click',close);
    document.addEventListener('keydown',e => { if(e.key==='Escape') close(); });
    document.getElementById('themeBtn')?.addEventListener('click',e => { e.preventDefault(); const theme=document.documentElement.dataset.theme==='dark'?'light':'dark'; document.documentElement.dataset.theme=theme; storage.set('lp_theme',theme); });
  });
})();
