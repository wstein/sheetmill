// SPDX-License-Identifier: EUPL-1.2
// Unit tests for the dependency-free search ranking in the theme's app.js.
import test from 'node:test';
import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import vm from 'node:vm';

const source = await readFile(new URL('../../internal/render/hugo/theme/static/assets/app.js', import.meta.url), 'utf8');
const context = {};
vm.runInNewContext(source, context); // without a document the script only defines its pure helpers
const { segments } = context.sheetmillSearch;
// Arrays from the sandbox have another realm's prototype; compare plain copies.
const rank = (items, query) => JSON.parse(JSON.stringify(context.sheetmillSearch.rank(items, query)));

const index = [
  { title: 'Undo changes', sheet: 'Git', section: 'History', text: 'reset and revert commits', tags: ['vcs'], category: 'version-control' },
  { title: 'Create an archive', sheet: 'tar', section: 'Basics', text: 'undo is not possible', tags: ['archive'], category: 'files' },
  { title: 'Revert a commit', sheet: 'Git', section: 'Undo', text: 'safe for shared history', tags: ['vcs'], category: 'version-control' },
];

test('every term must match somewhere', () => {
  assert.deepEqual(rank(index, 'undo nothing').map(i => i.title), []);
  assert.deepEqual(rank(index, '').map(i => i.title), []);
});
test('card titles outrank sections and text; ties keep index order', () => {
  assert.deepEqual(rank(index, 'undo').map(i => i.title), ['Undo changes', 'Revert a commit', 'Create an archive']);
});
test('tags and categories are searchable', () => {
  assert.deepEqual(rank(index, 'vcs revert').map(i => i.title), ['Revert a commit', 'Undo changes']);
  assert.deepEqual(rank(index, 'FILES').map(i => i.title), ['Create an archive']);
});
test('segments mark matches case-insensitively without overlaps', () => {
  assert.deepEqual(JSON.parse(JSON.stringify(segments('Revert a Commit', ['commit', 'rev', 'reve']))), [['Reve', true], ['rt a ', false], ['Commit', true]]);
  assert.deepEqual(JSON.parse(JSON.stringify(segments('plain', ['x']))), [['plain', false]]);
});
