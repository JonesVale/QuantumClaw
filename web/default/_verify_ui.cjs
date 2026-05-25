// Auto-verify UI: screenshots at multiple sizes + centering checks
const { chromium } = require('playwright');
const sizes = [
  { w: 375, label: 'mobile' },
  { w: 768, label: 'tablet' },
  { w: 1440, label: 'desktop' },
  { w: 2560, label: 'wide' }
];

async function verify(url = 'http://localhost:3666/enterprise') {
  console.log('=== UI VERIFY: ' + url + ' ===');
  let allPass = true;

  for (const s of sizes) {
    const b = await chromium.launch();
    const p = await b.newPage({ viewport: { width: s.w, height: 900 } });
    await p.goto(url, { waitUntil: 'networkidle', timeout: 30000 }).catch(() => {});
    await p.waitForTimeout(2000);

    // Screenshot
    await p.screenshot({ path: 'verify_' + s.label + '.png', fullPage: false });

    // Centering checks
    const r = await p.evaluate((vpw) => {
      const results = [];
      const checks = [
        { sel: 'h1', name: 'H1' },
        { sel: '[class*=rounded-full][class*=border-amber]', name: 'Tag' },
        { sel: '.qc-wrap', name: 'QcWrap' }
      ];
      for (const c of checks) {
        const el = document.querySelector(c.sel);
        if (!el) { results.push({ name: c.name, pass: false, reason: 'not found' }); continue; }
        const rect = el.getBoundingClientRect();
        const center = Math.round(rect.x + rect.width / 2);
        const diff = Math.abs(center - vpw / 2);
        // qc-wrap is centered by CSS margin, diff should be 0
        // h1/tag center should match vp center
        results.push({
          name: c.name,
          pass: diff <= 2,
          center, vp_center: vpw / 2, diff,
          x: Math.round(rect.x), w: Math.round(rect.width)
        });
      }
      return results;
    }, s.w);

    let pass = r.every(x => x.pass);
    if (!pass) allPass = false;
    console.log(s.label + ' (' + s.w + 'px): ' + (pass ? 'PASS' : 'FAIL'));
    r.forEach(x => console.log('  ' + (x.pass ? '✅' : '❌') + ' ' + x.name + ': center=' + x.center + ' vp=' + x.vp_center + (x.reason ? ' (' + x.reason + ')' : '')));
    await b.close();
  }

  console.log('=== RESULT: ' + (allPass ? 'ALL PASS ✅' : 'SOME FAILED ❌') + ' ===');
  return allPass;
}

verify().then(pass => process.exit(pass ? 0 : 1));
