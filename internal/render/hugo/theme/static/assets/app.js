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
  fetch('/search.json').then(response => { if (!response.ok) throw Error('Search unavailable'); return response.json(); }).then(index => {
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
    button.type = 'button'; button.className = 'copy'; button.textContent = 'Copy code';
    button.addEventListener('click', async () => {
      try { await navigator.clipboard.writeText(code.textContent); button.textContent = 'Copied'; }
      catch { button.textContent = 'Select code to copy'; }
      setTimeout(() => { button.textContent = 'Copy code'; }, 2000);
    });
    code.parentElement.append(button);
  });
})();
