// SPDX-License-Identifier: EUPL-1.2
import { readFile, lstat } from 'node:fs/promises';
import { join, isAbsolute } from 'node:path';
const slug=/^[a-z0-9]+(?:-[a-z0-9]+)*$/;
const nonempty=v=>typeof v==='string'&&v.trim().length>0;
export async function loadCorpus(root, data, selection='smill') {
  if(!['smill','all'].includes(selection))throw Error('invalid corpus selection; use smill or all');
  if(!Array.isArray(data)||!data.length||data.some(c=>!slug.test(c.id||''))||new Set(data.map(c=>c.id)).size!==data.length||!data.some(c=>c.id==='smill'))throw Error('invalid or duplicate corpus identities');
  const loaded=[];
  for(const item of data) {
    if(!nonempty(item.input)||isAbsolute(item.input)||item.input.includes('\\')||item.input.split('/').some(p=>!p||p==='.'||p==='..')||!item.input.endsWith('.md'))throw Error('invalid corpus input path');
    let selected=root;
    for(const part of item.input.split('/')) {
      selected=join(selected,part);const stat=await lstat(selected);
      if(stat.isSymbolicLink())throw Error('corpus symlinks are unsupported');
    }
    if(!(await lstat(selected)).isFile())throw Error('corpus input must be a regular file');
    if(!nonempty(item.title)||!nonempty(item.product?.name)||!/^[a-f0-9]{40}$/.test(item.product?.revision||'')||!nonempty(item.search?.query)||!nonempty(item.search?.label)||!slug.test(item.search?.card||'')||!slug.test(item.copy?.card||'')||!nonempty(item.copy?.text)||!nonempty(item.attribution?.license)||!nonempty(item.attribution?.author))throw Error('invalid corpus oracle');
    if(!Array.isArray(item.cards)||!item.cards.length||item.cards.some(c=>!slug.test(c))||new Set(item.cards).size!==item.cards.length||!item.cards.includes(item.search.card)||!item.cards.includes(item.copy.card)||!Array.isArray(item.crossLinks))throw Error('invalid corpus card identities');
    if(item.permalink&&(!item.cards.includes(item.permalink.card)||!nonempty(item.permalink.label)))throw Error('invalid corpus permalink');
    loaded.push({...item,sourcePath:selected,sourceBytes:await readFile(selected)});
  }
  for(const item of loaded)for(const link of item.crossLinks) {
    const target=loaded.find(c=>c.id===link.sheet);
    if(!nonempty(link.label)||!target?.cards.includes(link.card))throw Error('invalid corpus cross-card target');
  }
  const chosen=selection==='all'?loaded:loaded.filter(c=>c.id==='smill');
  for(const item of chosen)for(const link of item.crossLinks)if(!chosen.some(c=>c.id===link.sheet))throw Error('cross-card target is not selected');
  return chosen;
}
