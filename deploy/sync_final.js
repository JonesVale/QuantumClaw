const fs = require('fs');
const path = require('path');
const { Client } = require('ssh2');

const goCode = `package main
import (
  "context"
  "fmt"
  _ "github.com/quantumclaw/quantumclaw/model"
  "github.com/quantumclaw/quantumclaw/service"
)
func main() {
  if err := service.SyncPopularApps(context.Background()); err != nil {
    fmt.Println("ERR:", err)
  } else {
    fmt.Println("DONE")
  }
}
`;

const tmpFile = path.join(require('os').tmpdir(), 'sync_main.go');
fs.writeFileSync(tmpFile, goCode, 'utf8');

const conn = new Client();
conn.on('ready', () => {
  conn.sftp((err, sftp) => {
    if (err) { console.error('sftp:', err); conn.end(); return; }
    sftp.fastPut(tmpFile, '/opt/quantumclaw/cmd_sync/main.go', (e) => {
      if (e) { console.error('put:', e); conn.end(); return; }
      console.log('uploaded');
      sftp.end();
      conn.exec('cd /opt/quantumclaw && CGO_ENABLED=1 GOPROXY=https://goproxy.cn,direct go run ./cmd_sync/main.go 2>&1', (e2, stream) => {
        if (e2) { console.error(e2.message); conn.end(); return; }
        let out = '';
        stream.on('data', d => out += d.toString());
        stream.on('close', () => { console.log(out); fs.unlinkSync(tmpFile); conn.end(); });
      });
    });
  });
});
conn.on('error', e => { console.error(e.message); process.exit(1); });
conn.connect({ host: '122.51.221.43', port: 22, username: 'ubuntu', password: 'Ctji@2020', readyTimeout: 300000 });
