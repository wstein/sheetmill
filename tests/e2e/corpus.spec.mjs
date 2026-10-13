// SPDX-License-Identifier: EUPL-1.2
import { test } from './helpers/fixtures.mjs';
import { readerJourney } from './helpers/journey.mjs';
for(const id of ['bdock','whr'])test.describe(id,()=>{
  test.use({corpusID:id});
  test('complete reader journey and cross-card links',readerJourney);
});
