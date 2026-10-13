// SPDX-License-Identifier: EUPL-1.2
(() => {
  const root = document.documentElement;
  const theme = document.querySelector('#app-theme');
  let preference;
  try { preference = localStorage.getItem('sheetmill-theme'); } catch {}
  root.dataset.theme = preference === 'dark' || preference === 'light' ? preference : (matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light');
  theme.hidden = false;
  const label = () => { theme.textContent = root.dataset.theme === 'dark' ? 'Light mode' : 'Dark mode'; };
  label();
  theme.addEventListener('click', () => {
    root.dataset.theme = root.dataset.theme === 'dark' ? 'light' : 'dark';
    try { localStorage.setItem('sheetmill-theme', root.dataset.theme); } catch {}
    label();
  });
  const input = document.querySelector('#app-search');
  const results = document.querySelector('#app-results');
  const status = document.querySelector('#app-search-status');
  fetch(`${root.dataset.base || '/'}search.json`).then(response => { if (!response.ok) throw Error('Search unavailable'); return response.json(); }).then(index => {
    input.disabled = false;
    status.textContent = 'Search titles and card contents.';
    input.addEventListener('input', () => {
      const terms = input.value.toLocaleLowerCase().trim().split(/\s+/).filter(Boolean);
      results.replaceChildren();
      results.hidden = terms.length === 0;
      if (!terms.length) { status.textContent = 'Search titles and card contents.'; return; }
      const matches = index.filter(item => terms.every(term => `${item.title} ${item.sheet} ${item.section} ${item.text}`.toLocaleLowerCase().includes(term)));
      for (const item of matches.slice(0, 30)) {
        const li = document.createElement('li');
        const link = document.createElement('a');
        link.href = item.url;
        link.textContent = `${item.sheet} · ${item.title}`;
        li.append(link); results.append(li);
      }
      status.textContent = `${matches.length} results${matches.length > 30 ? ' (first 30 shown)' : ''}.`;
    });
  }).catch(() => { status.textContent = 'Search unavailable. Browse the library below.'; });
  if (navigator.clipboard?.writeText) document.querySelectorAll('pre > code').forEach(code => {
    const button = document.createElement('button');
    button.type = 'button'; button.className = 'copy';
    const icon = document.createElementNS('http://www.w3.org/2000/svg', 'svg');
    icon.setAttribute('viewBox', '0 0 24 24'); icon.setAttribute('width', '16'); icon.setAttribute('height', '16');
    icon.setAttribute('fill', 'none'); icon.setAttribute('stroke', 'currentColor'); icon.setAttribute('stroke-width', '1.7');
    icon.setAttribute('stroke-linecap', 'round'); icon.setAttribute('stroke-linejoin', 'round'); icon.setAttribute('aria-hidden', 'true');
    const path = document.createElementNS('http://www.w3.org/2000/svg', 'path'); icon.append(path);
    const feedback = document.createElement('span'); feedback.className = 'copy-feedback'; feedback.setAttribute('role', 'status');
    button.append(icon);
    const states = {
      idle: ['Copy code', 'M9 9h11v11H9z M5 15H4V4h11v1', ''],
      copied: ['Copied', 'm5 12 4 4L19 6', 'Code copied'],
      error: ['Select code to copy', 'M12 7v6m0 4h.01 M4 3h16v18H4z', 'Copy failed. Select code to copy.'],
    };
    const show = state => {
      const [label, shape, message] = states[state];
      button.setAttribute('aria-label', label); button.title = label;
      path.setAttribute('d', shape); feedback.textContent = message;
    };
    let reset;
    show('idle');
    button.addEventListener('click', async () => {
      clearTimeout(reset);
      try { await navigator.clipboard.writeText(code.textContent); show('copied'); }
      catch { show('error'); }
      reset = setTimeout(() => { show('idle'); }, 2000);
    });
    code.parentElement.insertBefore(button, code);
    // Keep the live region outside the button's presentational descendants.
    code.parentElement.insertBefore(feedback, code);
  });
})();
