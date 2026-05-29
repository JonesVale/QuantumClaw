const { Client } = require('ssh2');
const { execSync } = require('child_process');
const fs = require('fs');
const path = require('path');

const HOST = '122.51.221.43';
const USER = 'ubuntu';
const PASS = 'Ctji@2020';
const LOCAL_DIR = 'H:/AiData/openclaw/workspace/QuantumClaw';
const TARBALL = path.join(require('os').tmpdir(), 'qc-deploy.tar.gz');

console.log('Packing...');
execSync('tar --exclude=.git --exclude=node_modules --exclude=logs --exclude=bin --exclude=release --exclude=deploy --exclude=electron --exclude=data -czf "' + TARBALL + '" .', { cwd: LOCAL_DIR });
const sizeMB = (fs.statSync(TARBALL).size / 1024 / 1024).toFixed(1);
console.log('Tarball: ' + sizeMB + ' MB');

const conn = new Client();
conn.on('ready', () => {
  console.log('Uploading...');
  conn.sftp((err, sftp) => {
    if (err) { console.error(err.message); conn.end(); return; }
    sftp.fastPut(TARBALL, '/tmp/qc-deploy.tar.gz', (e) => {
      if (e) { console.error(e.message); conn.end(); return; }
      console.log('Uploaded');
      sftp.end();
      try { fs.unlinkSync(TARBALL); } catch(ex) {}
      console.log('Building...');
      conn.exec('killall -9 go quantumclaw 2>/dev/null; sleep 2; mkdir -p /opt/quantumclaw && cd /opt/quantumclaw && cp sqlite/quantumclaw.db /tmp/qc-db-backup 2>/dev/null; rm -rf * .??* 2>/dev/null; tar xzf /tmp/qc-deploy.tar.gz; cp /tmp/qc-db-backup sqlite/quantumclaw.db 2>/dev/null; rm /tmp/qc-db-backup 2>/dev/null; rm /tmp/qc-deploy.tar.gz; CGO_ENABLED=1 GOPROXY=https://goproxy.cn,direct go build -o quantumclaw . 2>&1 && echo BUILD_OK', (e2, stream) => {
        if (e2) { console.error(e2.message); conn.end(); return; }
        let out = '';
        stream.on('data', d => out += d.toString());
        stream.on('close', () => {
          if (out.includes('BUILD_OK')) {
            console.log('Starting...');
            conn.exec('nohup /opt/quantumclaw/quantumclaw > /opt/quantumclaw/quantumclaw.log 2>&1 & sleep 5; curl -s http://localhost:3666/api/status | head -c 80', (e3, s3) => {
              let o = '';
              s3.on('data', d => o += d.toString());
              s3.on('close', () => {
                console.log('Deployed. API: ' + o.substring(0,60));
                conn.end();
              });
            });
          } else { console.log(out); conn.end(); }
        });
      });
    });
  });
});
conn.on('error', e => { console.error(e.message); process.exit(1); });
conn.connect({ host: HOST, port: 22, username: USER, password: PASS, readyTimeout: 300000 });
