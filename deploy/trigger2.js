const { Client } = require('ssh2');
const conn = new Client();
conn.on('ready', () => {
  conn.exec('cd /opt/quantumclaw && killall -9 quantumclaw 2>/dev/null; sleep 1 && CGO_ENABLED=1 GOPROXY=https://goproxy.cn,direct go run main.go 2>&1 & BGPID=$!; sleep 25 && kill $BGPID 2>/dev/null; echo "=== 25 seconds should be enough for startup sync ==="', (e, s) => {
    if(e) { console.error(e); conn.end(); return; }
    s.on('data', d => process.stdout.write(d));
    s.on('close', () => conn.end());
  });
});
conn.on('error', e => { console.error(e.message); process.exit(1); });
conn.connect({ host:'122.51.221.43', port:22, username:'ubuntu', password:'Ctji@2020', readyTimeout:60000 });
