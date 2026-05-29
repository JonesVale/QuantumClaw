// QuantumClaw Deploy Script v6 — tarball 方式（跳过前端构建，使用本地 pre-built dist）
const { Client } = require('ssh2');
const { execSync } = require('child_process');
const fs = require('fs');
const path = require('path');

const HOST = '122.51.221.43';
const USER = 'ubuntu';
const PASS = 'Ctji@2020';
const PROJECT_DIR = '/opt/quantumclaw';
const LOCAL_DIR = 'H:\\AiData\\openclaw\\workspace\\QuantumClaw';
const TARBALL = path.join(process.env.TEMP || 'C:\\Temp', 'qc-deploy.tar.gz');

function execLocal(cmd) {
  console.log('  [local] ' + cmd.substring(0, 100));
  return execSync(cmd, { encoding: 'utf8', maxBuffer: 50 * 1024 * 1024, cwd: LOCAL_DIR, stdio: 'pipe' });
}

const conn = new Client();

function execRemote(cmd, timeout = 120000) {
  return new Promise((resolve, reject) => {
    console.log('  [remote] ' + cmd.substring(0, 100));
    conn.exec(cmd, (err, stream) => {
      if (err) return reject(err);
      let out = '';
      stream.on('data', (d) => out += d.toString());
      stream.stderr.on('data', (d) => process.stderr.write(d.toString()));
      stream.on('close', () => resolve(out));
    });
  });
}

conn.on('ready', async () => {
  console.log('✅ SSH connected to ' + HOST);
  try {
    console.log('\n📦 Packing source + pre-built dist...');
    // Only include Go source + pre-built dist (exclude node_modules, .git, etc.)
    execLocal('tar --exclude=.git --exclude=node_modules --exclude=logs --exclude=bin --exclude=release --exclude=deploy --exclude=electron --exclude=data --exclude="*.exe" --exclude="*.png" --exclude="*.jpg" -czf "' + TARBALL + '" .');
    const sizeMB = (fs.statSync(TARBALL).size / 1024 / 1024).toFixed(1);
    console.log('  Tarball: ' + sizeMB + ' MB');

    console.log('\n📤 Uploading...');
    const sftp = await new Promise((resolve, reject) => conn.sftp((err, s) => err ? reject(err) : resolve(s)));
    const remoteTarball = '/tmp/qc-deploy.tar.gz';
    await new Promise((resolve, reject) => sftp.fastPut(TARBALL, remoteTarball, (err) => err ? reject(err) : resolve()));
    console.log('  Upload complete');
    sftp.end();
    try { fs.unlinkSync(TARBALL); } catch(e) {}

    console.log('\n📂 Extracting...');
    await execRemote('mkdir -p ' + PROJECT_DIR + ' && cd ' + PROJECT_DIR + ' && sudo rm -rf * .??* 2>/dev/null && tar xzf ' + remoteTarball + ' && rm ' + remoteTarball + ' && echo EXTRACT_OK');

    console.log('\n🏗️  Building Go backend...');
    await execRemote('cd ' + PROJECT_DIR + ' && export CGO_ENABLED=1 && export GOPROXY=https://goproxy.cn,direct && go build -o quantumclaw . 2>&1 && echo BUILD_OK', 300000);

    console.log('\n🔄 Restarting...');
    await execRemote('sudo pkill -f quantumclaw 2>/dev/null; sleep 2');
    await execRemote('cd ' + PROJECT_DIR + ' && nohup ./quantumclaw > quantumclaw.log 2>&1 &');
    await new Promise(res => setTimeout(res, 5000));

    console.log('\n🏥 Health check...');
    let r = await execRemote('curl -s http://localhost:3666/api/status');
    console.log('  API Status: ' + r.substring(0, 200) + '...');

    console.log('\n✅ ========================================');
    console.log('   QuantumClaw Deployed!');
    console.log('   http://' + HOST);
    console.log('   ========================================');
    conn.end();
  } catch (err) {
    console.error('❌ Error:', err.message);
    conn.end();
  }
});

conn.on('error', (err) => console.error('SSH error:', err.message));
conn.connect({ host: HOST, port: 22, username: USER, password: PASS, readyTimeout: 15000 });
