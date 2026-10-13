// SPDX-License-Identifier: EUPL-1.2
(() => {
  'use strict';
  const chrome=document.createElement('aside');chrome.id='studio-status';chrome.setAttribute('aria-live','polite');
  const label=document.createElement('strong'),details=document.createElement('div'),links=document.createElement('nav');links.setAttribute('aria-label','Studio references');
  chrome.append(label,details,links);document.body.prepend(chrome);
  let state=null,connected=false;
  const selected=location.pathname.split('/')[1];
  let seen=Number(document.currentScript?.dataset.published||0);
  function render(){
    const sheet=state?.sheets.find(s=>s.id===selected);
    label.textContent=!connected?'Disconnected — freshness unknown':state?.error?'Preview failed':sheet?`${sheet.status}${sheet.stale?' — last-valid preview':''}`:'SheetMill Studio — read-only';
    details.replaceChildren();links.replaceChildren();
    if(state?.error){const p=document.createElement('p');p.textContent=state.error;details.append(p)}
    for(const s of state?.sheets||[]){
      const a=document.createElement('a');a.href='/'+s.id+'/';a.textContent=`${s.title||s.name} · ${s.status}${s.stale?' (last-valid)':''}`;links.append(a);
      for(const d of s.diagnostics||[]){const p=document.createElement('p');p.textContent=`${s.name}:${d.span.start.line}:${d.span.start.column}: ${d.code}: ${d.message}`;details.append(p)}
    }
  }
  async function refresh(){
    try{
      const response=await fetch('/studio/state',{cache:'no-store'});if(!response.ok)throw Error('State unavailable');
      const next=await response.json();if(state&&next.version<state.version)return;state=next;render();
      if(connected&&next.published>seen){seen=next.published;location.reload()}
    }catch{connected=false;render()}
  }
  let events;
  function connect(){
    events?.close();events=new EventSource('/studio/events');
    events.onopen=()=>{connected=true;refresh()};events.onmessage=()=>refresh();events.onerror=()=>{connected=false;render()};
  }
  window.addEventListener('offline',()=>{events?.close();connected=false;render()});
  window.addEventListener('online',()=>{connect();refresh()});connect();render();refresh();
})();
