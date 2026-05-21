const {Client} = require('ssh2');
const fs = require('fs');

const DB_PATH = 'H:\\AiData\\openclaw\\workspace\\QuantumClaw\\data\\quantumclaw.db';
const HOST = '139.196.8.90';
const USER = 'root';
const PASS = 'Jones.Vale@01';

const conn = new Client();

function execRemote(cmd, timeout = 30000) {
  return new Promise((resolve, reject) => {
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
  console.log('✅ SSH connected');
  try {
    // Step 1: Stop container
    console.log('\n🛑 Stopping cloud container...');
    let r = await execRemote('docker stop quantumclaw 2>/dev/null; echo STOPPED');

    // Step 2: Get volume mount path and upload DB
    console.log('\n📂 Finding volume mount...');
    r = await execRemote('docker volume inspect quantumclaw_quantumclaw_data --format "{{.Mountpoint}}" 2>/dev/null || echo /var/lib/docker/volumes/quantumclaw_quantumclaw_data/_data');
    const mountPath = r.out.trim() || '/var/lib/docker/volumes/quantumclaw_quantumclaw_data/_data';
    console.log('  Mount:', mountPath);

    // Step 3: Upload DB via SFTP
    console.log('\n📤 Uploading database (2.7 MB)...');
    const sftp = await new Promise((resolve, reject) => conn.sftp((err, s) => err ? reject(err) : resolve(s)));
    const stat = fs.statSync(DB_PATH);
    console.log(`  Local DB: ${(stat.size/1024/1024).toFixed(1)} MB`);

    await new Promise((resolve, reject) => {
      sftp.fastPut(DB_PATH, '/tmp/quantumclaw.db', (err) => err ? reject(err) : resolve());
    });
    console.log('  Upload complete');
    sftp.end();

    // Step 4: Copy to volume mount
    console.log('\n📋 Copying to volume...');
    r = await execRemote(`cp /tmp/quantumclaw.db ${mountPath}/quantumclaw.db && chown 1000:1000 ${mountPath}/quantumclaw.db && ls -la ${mountPath}/quantumclaw.db && rm /tmp/quantumclaw.db`);

    // Step 5: Restart container
    console.log('\n🚀 Starting container...');
    r = await execRemote('cd /opt/quantumclaw && docker compose -f docker-compose.sqlite.yml up -d 2>&1 || docker start quantumclaw 2>&1');

    // Step 6: Health check
    await new Promise(res => setTimeout(res, 5000));
    console.log('\n🏥 Health check...');
    r = await execRemote('docker ps --filter name=quantumclaw --format "{{.Status}}" 2>&1');
    console.log('  Status:', r.out.trim());
    r = await execRemote('docker logs --tail 3 quantumclaw 2>&1 || echo no_logs');

    console.log('\n✅ Database uploaded and container restarted!');
    conn.end();
  } catch (err) {
    console.error('❌', err.message);
    conn.end();
  }
});

conn.on('error', (err) => console.error('SSH error:', err.message));
conn.connect({ host: HOST, port: 22, username: USER, password: PASS, readyTimeout: 15000 });
