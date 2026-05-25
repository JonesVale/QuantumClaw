const { chromium } = require('playwright');
const sizes = [
  { w: 375, label: 'mobile' },
  { w: 768, label: 'tablet' },
  { w: 1440, label: 'desktop' },
  { w: 2560, label: 'wide' }
];

async function auditResponsive(url, name) {
  console.log('=== ' + name + ' @ ' + url + ' ===');
  let allPass = true;

  for (const s of sizes) {
    const b = await chromium.launch();
    const p = await b.newPage({ viewport: { width: s.w, height: 900 } });
    await p.goto(url, { waitUntil: 'networkidle', timeout: 30000 }).catch(() => {});
    await p.waitForTimeout(3000);

    const issues = await p.evaluate((vpw) => {
      const found = [];
      const body = document.body;
      const bw = body.scrollWidth;

      // 1) Horizontal overflow
      if (bw > vpw + 5) found.push('HORIZONTAL_SCROLL: body=' + bw + ' viewport=' + vpw);

      // 2) Center check on main containers
      const qc = document.querySelector('.qc-wrap');
      if (qc) {
        const r = qc.getBoundingClientRect();
        const diff = Math.abs(r.x + r.width / 2 - vpw / 2);
        if (diff > 3) found.push('QC_NOT_CENTERED: diff=' + diff);
      }

      // 3) Text overflow in any element
      document.querySelectorAll('p, h1, h2, h3, h4, button, span, a').forEach(el => {
        if (el.scrollWidth > el.clientWidth + 2) {
          const txt = (el.textContent || '').trim().slice(0, 40);
          if (txt) found.push('TEXT_OVERFLOW: "' + txt + '" overflow=' + (el.scrollWidth - el.clientWidth));
        }
      });

      // 4) Empty/collapsed sections
      document.querySelectorAll('section, div.grid > div, div.flex-1').forEach(el => {
        const txt = (el.textContent || '').trim();
        if (txt.length === 0 && el.children.length === 0) found.push('EMPTY_SECTION');
      });

      return found;
    }, s.w);

    if (issues.length > 0) {
      allPass = false;
      console.log(s.label + ' (' + s.w + 'px): FAIL');
      issues.forEach(i => console.log('  ❌ ' + i));
    } else {
      console.log(s.label + ' (' + s.w + 'px): PASS');
    }

    await p.screenshot({ path: name + '_' + s.label + '.png' });
    await b.close();
  }

  console.log(name + ': ' + (allPass ? 'ALL PASS ✅' : 'HAS ISSUES ❌'));
  return allPass;
}

(async () => {
  await auditResponsive('http://localhost:3666/models', 'models');
})();
