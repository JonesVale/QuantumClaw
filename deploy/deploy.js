// QuantumClaw Cloud Deploy Script v3 - tarball approach
// ⚠️ 编辑前: 设置下方 HOST/USER/PASS 为你的腾讯云服务器信息
const { Client } = require('ssh2');
const { execSync } = require('child_process');
const fs = require('fs');
const path = require('path');

const HOST = '122.51.221.43';  // ← 腾讯云服务器 IP
const USER = 'root';           // ← SSH 用户名
const PASS = '';               // ← SSH 密码（请填写）
const PROJECT_DIR = '/opt/quantumclaw';
const LOCAL_DIR = 'H:\\AiData\\openclaw\\workspace\\QuantumClaw';
const TARBALL = path.join(process.env.TEMP || 'C:\\Temp', 'quantumclaw-deploy.tar.gz');

function execLocal(cmd) {
  console.log(`  [local] ${cmd.substring(0, 80)}`);
  return execSync(cmd, { encoding: 'utf8', maxBuffer: 10 * 1024 * 1024, cwd: LOCAL_DIR, stdio: 'pipe' });
}

const conn = new Client();

function execRemote(cmd, timeout = 120000) {
  return new Promise((resolve, reject) => {
    console.log(`  [remote] ${cmd.substring(0, 100)}`);
    conn.exec(cmd, (err, stream) => {
      if (err) return reject(err);
      let out = '', errOut = '';
      stream.on('data', (d) => { out += d.toString(); process.stdout.write(d.toString()); });
      stream.stderr.on('data', (d) => { errOut += d.toString(); });
      stream.on('close', () => resolve({ out, err: errOut }));
    });
  });
}

conn.on('ready', async () => {
  console.log('✅ SSH connected to ' + HOST);
  try {
    // Step 1: Check server
    let r = await execRemote('uname -a; docker --version 2>&1; docker compose version 2>&1; echo ---READY---');
    
    // Step 2: Create tarball locally
    console.log('\n📦 Creating tarball (excluding .git/node_modules/logs/bin/electron)...');
    execLocal('tar --exclude=.git --exclude=node_modules --exclude=logs --exclude=bin --exclude=release --exclude=deploy --exclude=electron --exclude=data -czf "' + TARBALL + '" .');
    const sizeMB = (fs.statSync(TARBALL).size / 1024 / 1024).toFixed(1);
    console.log(`  Tarball: ${sizeMB} MB`);

    // Step 3: Upload tarball via SFTP
    console.log('\n📤 Uploading tarball...');
    const sftp = await new Promise((resolve, reject) => conn.sftp((err, s) => err ? reject(err) : resolve(s)));
    const tarballRemote = '/tmp/quantumclaw-deploy.tar.gz';
    await new Promise((resolve, reject) => {
      sftp.fastPut(TARBALL, tarballRemote, (err) => err ? reject(err) : resolve());
    });
    console.log('  Upload complete');
    sftp.end();
    
    try { fs.unlinkSync(TARBALL); } catch(e) {}

    // Step 4: Extract on server and build
    console.log('\n📂 Extracting on server...');
    r = await execRemote(`mkdir -p ${PROJECT_DIR} && cd ${PROJECT_DIR} && tar xzf ${tarballRemote} && rm ${tarballRemote} && echo EXTRACT_OK`);
    
    // Step 5: Docker compose down + up --build
    console.log('\n🛑 Stopping old containers...');
    await execRemote(`cd ${PROJECT_DIR} && docker compose -f docker-compose.sqlite.yml down 2>/dev/null; echo ok`);
    
    console.log('\n🔨 Building & starting Docker (2-4 min)...');
    r = await execRemote(`cd ${PROJECT_DIR} && docker compose -f docker-compose.sqlite.yml up -d --build 2>&1`, 300000);

    // Step 6: Health check
    await new Promise(res => setTimeout(res, 10000));
    console.log('\n🏥 Health check...');
    r = await execRemote('docker ps --filter name=quantumclaw --format "{{.Status}}" 2>&1 || echo NOT_RUNNING');
    r = await execRemote('docker logs --tail 8 quantumclaw 2>&1 || echo NO_LOGS');
    
    // Step 7: Firewall
    await execRemote('(ufw allow 3666/tcp 2>/dev/null || firewall-cmd --add-port=3666/tcp --permanent 2>/dev/null && firewall-cmd --reload 2>/dev/null || iptables -I INPUT -p tcp --dport 3666 -j ACCEPT 2>/dev/null) && echo FIREWALL_OK || echo FW_SKIP');

    console.log(`\n✅ ========================================`);
    console.log(`   QuantumClaw Deployed!`);
    console.log(`   http://${HOST}`);
    console.log(`   ========================================`);
    conn.end();
  } catch (err) {
    console.error('❌ Error:', err.message);
    conn.end();
  }
});

conn.on('error', (err) => console.error('SSH error:', err.message));
conn.connect({ host: HOST, port: 22, username: USER, password: PASS, readyTimeout: 15000 });
