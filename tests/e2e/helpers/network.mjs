// SPDX-License-Identifier: EUPL-1.2
export async function guardNetwork(context,origin) {
  const violations=[];
  await context.route('**/*',async route=>{
    const url=route.request().url();
    if(new URL(url).origin===origin)await route.continue();
    else {violations.push(url);await route.abort('blockedbyclient');}
  });
  return violations;
}
