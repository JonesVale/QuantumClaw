// QuantumClaw - Upload local SQLite DB to Remote Server
// ⚠️ 编辑前: 设置下方 HOST/PASS 为你的腾讯云服务器信息
const {Client} = require('ssh2');
const fs = require('fs');

const DB_PATH = 'H:\\AiData\\openclaw\\workspace\\QuantumClaw\\data\\quantumclaw.db';
const HOST = '122.51.221.43';
const USER = 'root';
const PASS = '';

const conn = new Client();


conn.on('ready', async () => {
  console.log('✅ SSH connected');
  try {
    const sftp = await new Promise((resolve, reject) => conn.sftp((err, s) => err ? reject(err) : resolve(s)));
    const remotePath = '/opt/quantumclaw/data/quantumclaw.db';
    
    console.log(`Uploading ${DB_PATH} → ${remotePath} ...`);
    await new Promise((resolve, reject) => {
      sftp.fastPut(DB_PATH, remotePath, (err) => err ? reject(err) : resolve());
    });
    console.log('Upload complete');
    sftp.end();
    conn.end();
  } catch (err) {
    console.error('Error:', err.message);
    conn.end();
  }
});

conn.on('error', (err) => console.error('SSH error:', err.message));
conn.connect({ host: HOST, port: 22, username: USER, password: PASS, readyTimeout: 15000 });
