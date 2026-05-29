const { Client } = require('ssh2');
const fs = require('fs');
const conn = new Client();
conn.on('ready', () => {
  conn.sftp((err, sftp) => {
    if (err) { console.error(err); conn.end(); return; }
    // Upload changed files
    sftp.fastPut('H:\\AiData\\openclaw\\workspace\\QuantumClaw\\controller\\app_market.go', '/opt/quantumclaw/controller/app_market.go', (e) => {
      if (e) { console.error(e.message); conn.end(); return; }
      sftp.fastPut('H:\\AiData\\openclaw\\workspace\\QuantumClaw\\router\\api.go', '/opt/quantumclaw/router/api.go', (e2) => {
        if (e2) { console.error(e2.message); conn.end(); return; }
        sftp.end();
        console.log("Uploaded. Rebuilding...");
        conn.exec("cd /opt/quantumclaw && export CGO_ENABLED=1 GOPROXY=https://goproxy.cn,direct && go build -o quantumclaw . 2>&1 && echo BUILD_OK && killall quantumclaw 2>/dev/null; sleep 2 && nohup ./quantumclaw > quantumclaw.log 2>&1 & echo RESTARTED", (e3, stream) => {
          if (e3) { console.error(e3.message); conn.end(); return; }
          let out = "";
          stream.on("data", d => out += d.toString());
          stream.on("close", () => { console.log(out); conn.end(); });
        });
      });
    });
  });
});
conn.on("error", e => { console.error(e.message); process.exit(1); });
conn.connect({ host:"122.51.221.43", port:22, username:"ubuntu", password:"Ctji@2020", readyTimeout:60000 });
