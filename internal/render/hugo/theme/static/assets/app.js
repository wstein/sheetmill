// SPDX-License-Identifier: EUPL-1.2
// Pure search helpers: every term must match; fields are weighted so card
// titles outrank sheet titles, tags/categories, sections and body text.
const sheetmillSearch = (() => {
  const fields = [['title', 10], ['sheet', 6], ['labels', 5], ['section', 3], ['text', 1]];
  const terms = query => query.toLocaleLowerCase().trim().split(/\s+/).filter(Boolean);
  const rank = (index, query) => {
    const wanted = terms(query);
    if (!wanted.length) return [];
    const scored = [];
    index.forEach((item, position) => {
      const values = { title: item.title, sheet: item.sheet, labels: [...(item.tags || []), item.category || ''].join(' '), section: item.section, text: item.text };
      let score = 0;
      for (const term of wanted) {
        let best = 0;
        for (const [field, weight] of fields) {
          const value = (values[field] || '').toLocaleLowerCase();
          const at = value.indexOf(term);
          if (at < 0) continue;
          const wordStart = at === 0 || /[^\p{L}\p{N}]/u.test(value[at - 1]);
          best = Math.max(best, weight + (wordStart ? 1 : 0));
        }
        if (!best) return;
        score += best;
      }
      scored.push({ item, score, position });
    });
    return scored.sort((a, b) => b.score - a.score || a.position - b.position).map(s => s.item);
  };
  // segments splits text into [part, matched] pairs; earliest, then longest match wins.
  const segments = (text, wanted) => {
    const lower = text.toLocaleLowerCase(), out = [];
    let i = 0, plain = '';
    while (i < text.length) {
      let length = 0;
      for (const term of wanted) if (term && lower.startsWith(term, i)) length = Math.max(length, term.length);
      if (length) {
        if (plain) { out.push([plain, false]); plain = ''; }
        out.push([text.slice(i, i + length), true]);
        i += length;
      } else plain += text[i++];
    }
    if (plain) out.push([plain, false]);
    return out;
  };
  return { rank, segments, terms };
})();
globalThis.sheetmillSearch = sheetmillSearch;
if (typeof document !== 'undefined') (() => {
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
  const idle = 'Search titles, tags and card contents. Press / to search.';
  const links = () => [...results.querySelectorAll('a')];
  // "/" or Ctrl/Cmd+K focuses search unless the reader is already typing.
  document.addEventListener('keydown', event => {
    const typing = /^(INPUT|TEXTAREA|SELECT)$/.test(document.activeElement?.tagName) || document.activeElement?.isContentEditable;
    if ((event.key === '/' && !typing) || ((event.ctrlKey || event.metaKey) && event.key.toLowerCase() === 'k')) {
      if (input.disabled) return;
      event.preventDefault(); input.focus(); input.select();
    }
  });
  input.addEventListener('keydown', event => {
    if (event.key === 'ArrowDown' && links().length) { event.preventDefault(); links()[0].focus(); }
    else if (event.key === 'Enter' && links().length) { event.preventDefault(); links()[0].click(); }
    else if (event.key === 'Escape') { input.value = ''; input.dispatchEvent(new Event('input')); }
  });
  results.addEventListener('keydown', event => {
    const all = links(), at = all.indexOf(document.activeElement);
    if (event.key === 'ArrowDown' && at < all.length - 1) { event.preventDefault(); all[at + 1].focus(); }
    else if (event.key === 'ArrowUp') { event.preventDefault(); (at > 0 ? all[at - 1] : input).focus(); }
    else if (event.key === 'Escape') { event.preventDefault(); input.focus(); }
  });
  const highlighted = (text, wanted) => sheetmillSearch.segments(text, wanted).map(([part, match]) => {
    if (!match) return document.createTextNode(part);
    const mark = document.createElement('mark'); mark.textContent = part; return mark;
  });
  fetch(`${root.dataset.base || '/'}search.json`).then(response => { if (!response.ok) throw Error('Search unavailable'); return response.json(); }).then(index => {
    input.disabled = false;
    status.textContent = idle;
    input.addEventListener('input', () => {
      const wanted = sheetmillSearch.terms(input.value);
      results.replaceChildren();
      results.hidden = wanted.length === 0;
      if (!wanted.length) { status.textContent = idle; return; }
      const matches = sheetmillSearch.rank(index, input.value);
      for (const item of matches.slice(0, 30)) {
        const li = document.createElement('li');
        const link = document.createElement('a');
        link.href = item.url;
        link.append(...highlighted(`${item.sheet} · ${item.title}`, wanted));
        li.append(link); results.append(li);
      }
      status.textContent = `${matches.length} results${matches.length > 30 ? ' (first 30 shown)' : ''}.${matches.length ? ' Use the arrow keys to move between them.' : ''}`;
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
