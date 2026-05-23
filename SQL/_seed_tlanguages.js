// T_Languages auto-seed: scan TSX/TS files for all translatable text,
// save to DB via API. Run once after frontend build.

const fs = require('fs');
const path = require('path');

const SRC = 'H:/AiData/openclaw/workspace/QuantumClaw/web/default/src';
const API = 'http://127.0.0.1:3666';

async function main() {
  // Step 1: Collect all unique t() keys from frontend source
  const keys = new Set();
  const files = walkDir(SRC);
  for (const f of files) {
    const c = fs.readFileSync(f, 'utf-8');
    const matches = c.matchAll(/t\(['"`]([^'"`]+)['"`]\)/g);
    for (const m of matches) {
      keys.add(m[1]);
    }
  }
  const sorted = [...keys].sort();
  console.log(`Found ${sorted.length} unique translation keys`);

  // Step 2: Generate English translations (key = display)
  const enEntries = sorted.map(key => ({ lcode: key, display: key, fromname: 'auto-seed' }));

  // Step 3: Check what we have for Chinese
  // Load zh-CN.json if readable
  let zhMap = {};
  try {
    // Try reading zh-CN.json - may be corrupted
    const zhRaw = fs.readFileSync(SRC + '/i18n/zh-CN.json', 'utf-8');
    // Try to extract key-value pairs using regex (since JSON may be garbled)
    const kvMatches = zhRaw.matchAll(/"([^"]+)":\s*"([^"]+)"/g);
    for (const m of kvMatches) {
      zhMap[m[1]] = m[2];
    }
    console.log(`Extracted ${Object.keys(zhMap).length} CN entries from JSON file`);
  } catch (e) {
    console.log('zh-CN.json not available, using keys as fallback');
  }

  // For keys not in zhMap, use key as fallback
  const cnEntries = sorted.map(key => ({
    lcode: key,
    display: zhMap[key] || key,
    fromname: 'auto-seed'
  }));

  // Step 4: Seed to DB via API
  const results = [];
  for (const [langType, entries] of [['English', enEntries], ['中文简体', cnEntries]]) {
    // Try public endpoint first
    let ok = await seedApi(langType, entries, '/api/languages/seed-public');
    if (!ok) {
      // Public may be restricted; try admin-protected (may fail without token)
      ok = await seedApi(langType, entries, '/api/languages/seed', true);
    }
    results.push({ lang: langType, success: ok, count: entries.length });
  }

  console.log('\nSeed results:', JSON.stringify(results, null, 2));
}

async function seedApi(langType, entries, endpoint, requireAdmin = false) {
  try {
    const headers = { 'Content-Type': 'application/json' };
    if (requireAdmin) {
      // Try with session cookie (may work in local Docker)
      headers['Cookie'] = '';
    }
    const r = await fetch(API + endpoint, {
      method: 'POST',
      headers,
      body: JSON.stringify({ languages_type: langType, entries }),
    });
    const data = await r.json();
    if (data.success) {
      console.log(`✅ ${langType}: seeded ${data.count} entries`);
      return true;
    }
    console.log(`❌ ${langType}: ${data.message || 'unknown error'}`);
    return false;
  } catch (e) {
    console.log(`❌ ${langType}: ${e.message}`);
    return false;
  }
}

function walkDir(dir) {
  const results = [];
  for (const entry of fs.readdirSync(dir, { withFileTypes: true })) {
    const full = path.join(dir, entry.name);
    if (entry.isDirectory() && entry.name !== 'node_modules') {
      results.push(...walkDir(full));
    } else if (entry.isFile() && /\.(tsx|ts)$/.test(entry.name)) {
      results.push(full);
    }
  }
  return results;
}

main().catch(console.error);
