const fs = require('fs');
const http = require('http');

// Complete all missing keys to all language JSON files
function rj(fp) { let r=fs.readFileSync(fp,'utf-8'); if(r.charCodeAt(0)===0xFEFF)r=r.slice(1); return JSON.parse(r); }
function wj(fp,d) { fs.writeFileSync(fp, JSON.stringify(d,null,'  '),'utf-8'); }

// The 145 missing keys from code
// API paths (83) - keep as-is since they're URLs
const apiPaths = [
  '/api/about','/api/admin/channel-affinity','/api/admin/channel-affinity/cache/stats',
  '/api/admin/custom-oauth','/api/admin/menus','/api/admin/performance',
  '/api/admin/promo-ads','/api/admin/resellers','/api/admin/subscription/plans',
  '/api/admin/task','/api/admin/task/poll','/api/admin/upstream',
  '/api/admin/withdrawals','/api/admin/withdrawals?status=pending',
  '/api/channel/models','/api/channel/test','/api/channel/types',
  '/api/channel/update_balance','/api/commission/self/records','/api/commission/setting',
  '/api/distributor/self','/api/enterprise-clients','/api/group','/api/home_page_content',
  '/api/languages','/api/log/self/stat','/api/log/stat','/api/metrics',
  '/api/model-catalog','/api/models','/api/models/rankings','/api/models/sync',
  '/api/notice','/api/oauth/telegram/widget','/api/option','/api/platform/config',
  '/api/promo-ads','/api/quantum/backends','/api/quantum/providers',
  '/api/reseller/balance','/api/reseller/stats','/api/reseller/stats?days=30',
  '/api/reseller/withdraw','/api/rss/articles','/api/settlement/config',
  '/api/site-content','/api/site-features','/api/site-providers','/api/site-stats',
  '/api/status','/api/subscription/plans','/api/task','/api/user/2fa',
  '/api/user/2fa/init','/api/user/logout','/api/user/self','/api/user/self/balance',
  '/api/user/self/billing/records','/api/user/self/billing/stats',
  '/api/user/self/checkin','/api/user/self/checkin/history',
  '/api/user/self/dashboard','/api/user/self/notifications',
  '/api/user/self/notifications/read_all','/api/user/self/notifications/unread_count',
  '/api/user/self/security/activity','/api/user/self/subscription/plans',
  '/api/user/self/subscription/self','/api/user/self/team',
  '/api/user/self/topup/info','/api/user/self/topup/list',
  '/api/user/self/transaction_logs','/api/user/self/upgrade',
  '/api/user/self/webauthn/credentials','/api/user/self/withdraw/available',
  '/api/user/self/withdraw/earnings','/api/user/self/withdraw/list',
  '/api/user/upgrade','/api/user/withdrawals',
  '/api/webauthn/register/begin',
  // Non-API special keys that should exist in JSON
  '@/i18n/en.json','@/i18n/zh-CN.json','@/stores/auth-store','Admin:','Email:','Site:',
  'test content',';','In','Out','3-12 characters'
];

// Display text keys (62) - need proper English text
const displayTexts = {
  'AI Model Catalog': 'AI Model Catalog',
  'Access your QuantumClaw dashboard': 'Access your QuantumClaw dashboard',
  'Add Item': 'Add Item',
  'Admin Account': 'Admin Account',
  'Admin Email': 'Admin Email',
  'Admin Username': 'Admin Username',
  'All Models': 'All Models',
  'All Quantum': 'All Quantum',
  'All Set!': 'All Set!',
  'Allow password-based registration': 'Allow password-based registration',
  'Allow user registration': 'Allow user registration',
  'Back': 'Back',
  'Basic Settings': 'Basic Settings',
  'Browse': 'Browse',
  'Categories': 'Categories',
  'Chat with local model...': 'Chat with local model...',
  'Chat with your local Ollama models. Select a model to begin.': 'Chat with your local Ollama models. Select a model to begin.',
  'Checking...': 'Checking...',
  'Click me': 'Click me',
  'Close sidebar': 'Close sidebar',
  'Confirm Password': 'Confirm Password',
  'Custom error UI': 'Custom error UI',
  'Description': 'Description',
  'Enter admin username': 'Enter admin username',
  'Finish': 'Finish',
  'Finishing...': 'Finishing...',
  'Local': 'Local',
  'Login failed': 'Login failed',
  'Network error': 'Network error',
  'No local models found': 'No local models found',
  'No model selected': 'No model selected',
  'Nothing here yet': 'Nothing here yet',
  'Quantum Providers': 'Quantum Providers',
  'QuantumClaw AI Chat': 'QuantumClaw AI Chat',
  'Reasoning & Logic': 'Reasoning & Logic',
  'Registration:': 'Registration:',
  'Remote': 'Remote',
  'Select a model from the catalog and start a conversation.': 'Select a model from the catalog and start a conversation.',
  'Select a model...': 'Select a model...',
  'Select local model...': 'Select local model...',
  'Setup completed! Redirecting...': 'Setup completed! Redirecting...',
  'Setup failed': 'Setup failed',
  'Show less': 'Show less',
  'Show more...': 'Show more...',
  'Show sidebar': 'Show sidebar',
  'Site Name': 'Site Name',
  'Something broke!': 'Something broke!',
  'Summary': 'Summary',
  'The name displayed on your platform': 'The name displayed on your platform',
  'Thinking...': 'Thinking...',
  'Type a message... (Enter to send, Shift+Enter for new line)': 'Type a message... (Enter to send, Shift+Enter for new line)',
  'View Rankings': 'View Rankings',
  'Welcome to QuantumClaw': 'Welcome to QuantumClaw',
  'admin@example.com': 'admin@example.com',
};

// Chinese translations for display texts
const zhTranslations = {
  'AI Model Catalog': 'AI 模型目录',
  'Access your QuantumClaw dashboard': '访问您的 QuantumClaw 仪表盘',
  'Add Item': '添加项目',
  'Admin Account': '管理员账户',
  'Admin Email': '管理员邮箱',
  'Admin Username': '管理员用户名',
  'All Models': '全部模型',
  'All Quantum': '全部量子',
  'All Set!': '全部就绪！',
  'Allow password-based registration': '允许密码注册',
  'Allow user registration': '允许用户注册',
  'Back': '返回',
  'Basic Settings': '基础设置',
  'Browse': '浏览',
  'Categories': '分类',
  'Chat with local model...': '与本地模型聊天...',
  'Chat with your local Ollama models. Select a model to begin.': '与本地 Ollama 模型聊天。选择一个模型开始。',
  'Checking...': '检查中...',
  'Click me': '点击我',
  'Close sidebar': '关闭侧边栏',
  'Confirm Password': '确认密码',
  'Custom error UI': '自定义错误界面',
  'Description': '描述',
  'Enter admin username': '输入管理员用户名',
  'Finish': '完成',
  'Finishing...': '完成中...',
  'Local': '本地',
  'Login failed': '登录失败',
  'Network error': '网络错误',
  'No local models found': '未找到本地模型',
  'No model selected': '未选择模型',
  'Nothing here yet': '暂无内容',
  'Quantum Providers': '量子供应商',
  'QuantumClaw AI Chat': 'QuantumClaw AI 聊天',
  'Reasoning & Logic': '推理与逻辑',
  'Registration:': '注册：',
  'Remote': '远程',
  'Select a model from the catalog and start a conversation.': '从目录中选择一个模型开始对话。',
  'Select a model...': '选择一个模型...',
  'Select local model...': '选择本地模型...',
  'Setup completed! Redirecting...': '设置完成！正在跳转...',
  'Setup failed': '设置失败',
  'Show less': '收起',
  'Show more...': '展开更多...',
  'Show sidebar': '显示侧边栏',
  'Site Name': '站点名称',
  'Something broke!': '出错了！',
  'Summary': '摘要',
  'The name displayed on your platform': '在您平台上显示的名称',
  'Thinking...': '思考中...',
  'Type a message... (Enter to send, Shift+Enter for new line)': '输入消息...(Enter 发送, Shift+Enter 换行)',
  'View Rankings': '查看排行榜',
  'Welcome to QuantumClaw': '欢迎使用 QuantumClaw',
  'admin@example.com': 'admin@example.com',
};

// Build complete missing data
const allMissing = {};
for (const k of apiPaths) allMissing[k] = k;
for (const [k, v] of Object.entries(displayTexts)) allMissing[k] = v;

// Get Chinese for API paths (same as key)
const zhMissing = {};
for (const k of apiPaths) zhMissing[k] = k;
for (const [k, v] of Object.entries(zhTranslations)) zhMissing[k] = v;

console.log('Missing keys to add: ' + Object.keys(allMissing).length);
console.log('  API paths: ' + apiPaths.length);
console.log('  Display texts: ' + Object.keys(displayTexts).length);

// Update en.json
const en = rj('web/default/src/i18n/en.json');
for (const [k,v] of Object.entries(allMissing)) en[k] = v;
wj('web/default/src/i18n/en.json', en);
console.log('en.json: ' + Object.keys(en).length + ' keys');

// Update zh-CN.json
const zh = rj('web/default/src/i18n/zh-CN.json');
for (const [k,v] of Object.entries(zhMissing)) zh[k] = v;
wj('web/default/src/i18n/zh-CN.json', zh);
console.log('zh-CN.json: ' + Object.keys(zh).length + ' keys');

// Update all other languages with English placeholders
const langList = ['zh-TW','fr','ja','vi','ru','es','de','ko','it','pt','ar','hi','id','th','nl','tr'];
for (const lk of langList) {
  const fp = 'web/default/src/i18n/' + lk + '.json';
  const data = rj(fp);
  for (const k of Object.keys(allMissing)) {
    if (data[k] === undefined) data[k] = allMissing[k];
  }
  wj(fp, data);
}
console.log('Other languages updated: ' + langList.length);

// Seed to DB
console.log('\nSeeding to DB...');
const DB_NAMES = {'en':'English','zh-CN':'中文简体','zh-TW':'中文繁体','fr':'Francais','ja':'Japanese','vi':'Tieng Viet','ru':'Russkiy','es':'Espanol','de':'Deutsch','ko':'Hangug-eo','it':'Italiano','pt':'Portugues','ar':'Arabic','hi':'Hindi','id':'Bahasa Indonesia','th':'Thai','nl':'Nederlands','tr':'Turkce'};

function login() { return new Promise(r=>{const j=JSON.stringify({username:'root',password:'123456'});const o={hostname:'localhost',port:3666,path:'/api/user/login',method:'POST',headers:{'Content-Type':'application/json','Content-Length':Buffer.byteLength(j)}};const q=http.request(o,res=>{let d='';res.on('data',c=>d+=c);res.on('end',()=>{const m=(res.headers['set-cookie']||[]).join(';').match(/session=([^;]+)/);r(m?m[1]:null);});});q.write(j);q.end();}); }
function post(url,d,c) { return new Promise(r=>{const j=JSON.stringify(d);const o={hostname:'localhost',port:3666,path:url,method:'POST',headers:{'Content-Type':'application/json','Content-Length':Buffer.byteLength(j)}};if(c)o.headers['Cookie']='session='+c;const q=http.request(o,res=>{let b='';res.on('data',c=>b+=c);res.on('end',()=>r(JSON.parse(b)));});q.write(j);q.end();}); }

(async () => {
  const cookie = await login();
  for (const [lk, db] of Object.entries(DB_NAMES)) {
    const fp = 'web/default/src/i18n/' + lk + '.json';
    const data = rj(fp);
    const entries = Object.entries(data)
      .filter(([k,v]) => k && v && k.length <= 500)
      .map(([k,v]) => ({lcode:k, display:v, fromname:'fix'}));
    let ok = 0;
    for (let i = 0; i < entries.length; i += 100) {
      const b = entries.slice(i, i+100);
      const r = await post('/api/languages/seed', {languages_type:db, entries:b}, cookie);
      ok += r.count || b.length;
    }
    console.log('  ' + db + ': ' + entries.length + ' -> ' + ok);
  }
  console.log('\nDone!');
})();
