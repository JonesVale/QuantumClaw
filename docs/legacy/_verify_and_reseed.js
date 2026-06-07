const http = require('http');
const fs = require('fs');
function rj(fp) {let r=fs.readFileSync(fp,'utf-8');if(r.charCodeAt(0)===0xFEFF)r=r.slice(1);return JSON.parse(r);}
function login() { return new Promise(r=>{const j=JSON.stringify({username:'root',password:'123456'});const o={hostname:'localhost',port:3666,path:'/api/user/login',method:'POST',headers:{'Content-Type':'application/json','Content-Length':Buffer.byteLength(j)}};const q=http.request(o,res=>{let d='';res.on('data',c=>d+=c);res.on('end',()=>{const m=(res.headers['set-cookie']||[]).join(';').match(/session=([^;]+)/);r(m?m[1]:null);});});q.write(j);q.end();}); }
function getTranslations(lang) { return new Promise(r => http.get('http://localhost:3666/api/translations?lang='+encodeURIComponent(lang), res => { let d='';res.on('data',c=>d+=c);res.on('end',()=>r(JSON.parse(d))); })); }

const DB_NAMES = {'en':'English','zh-CN':'中文简体','zh-TW':'中文繁体','fr':'Francais','ja':'Japanese','vi':'Tieng Viet','ru':'Russkiy','es':'Espanol','de':'Deutsch','ko':'Hangug-eo','it':'Italiano','pt':'Portugues','ar':'Arabic','hi':'Hindi','id':'Bahasa Indonesia','th':'Thai','nl':'Nederlands','tr':'Turkce'};

(async () => {
  console.log('=== DB 各语言数据量 ===');
  const needsReseed = [];
  for (const [lk, db] of Object.entries(DB_NAMES)) {
    const r = await getTranslations(db);
    const count = r.data ? Object.keys(r.data).length : 0;
    const fp = 'web/default/src/i18n/' + lk + '.json';
    const jsonCount = Object.keys(rj(fp)).length;
    const status = count >= jsonCount ? 'OK' : 'NEED_' + (jsonCount-count) + '_more';
    console.log('  ' + lk.padEnd(8) + ' DB:' + count + ' JSON:' + jsonCount + ' ' + status);
    if (count < jsonCount) needsReseed.push(lk);
  }
  
  if (needsReseed.length > 0) {
    console.log('\n=== Re-seeding ' + needsReseed.length + ' languages ===');
    const cookie = await login();
    for (const lk of needsReseed) {
      const db = DB_NAMES[lk];
      const fp = 'web/default/src/i18n/' + lk + '.json';
      const data = rj(fp);
      const entries = Object.entries(data).filter(([k,v]) => k && v && k.length <= 500).map(([k,v]) => ({lcode:k, display:v, fromname:'check'}));
      let ok = 0;
      for (let i = 0; i < entries.length; i += 100) {
        const b = entries.slice(i, i+100);
        const r = await new Promise(rr => {const j=JSON.stringify({languages_type:db, entries:b});const o={hostname:'localhost',port:3666,path:'/api/languages/seed',method:'POST',headers:{'Content-Type':'application/json','Content-Length':Buffer.byteLength(j),'Cookie':'session='+cookie}};const q=http.request(o,res=>{let d='';res.on('data',c=>d+=c);res.on('end',()=>rr(JSON.parse(d)));});q.write(j);q.end();});
        ok += r.count || b.length;
      }
      console.log('  ' + lk + ': ' + entries.length + ' -> ' + ok + ' (batch 100)');
    }
    console.log('Re-seed done');
  } else {
    console.log('\nAll languages fully seeded!');
  }
})();
