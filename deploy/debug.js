const {Client} = require('ssh2');
const conn = new Client();
conn.on('ready', () => {
  conn.exec('sqlite3 /var/lib/docker/volumes/quantumclaw_quantumclaw_data/_data/quantumclaw.db "SELECT username,role,status,length(password),substr(password,1,50) FROM users LIMIT 5;" 2>&1 || echo "no sqlite3 on host"', (err, stream) => {
    if(err) {console.error(err); conn.end(); return;}
    stream.on('data', d => process.stdout.write(d.toString()));
    stream.stderr.on('data', d => process.stderr.write(d.toString()));
    stream.on('close', () => conn.end());
  });
});
conn.on('error', e => console.error('ERR:', e.message));
conn.connect({host:'139.196.8.90',port:22,username:'root',password:'Jones.Vale@01',readyTimeout:10000});
