const { Client } = require("ssh2");
const conn = new Client();
conn.on("ready", () => {
  conn.exec("killall -9 go 2>/dev/null; sleep 2; echo KILLED", (e, s) => {
    if(e) { console.error(e); conn.end(); return; }
    let out = "";
    s.on("data", d => out += d.toString());
    s.on("close", () => { console.log(out); 
      conn.exec("cd /opt/quantumclaw && export CGO_ENABLED=1 GOPROXY=https://goproxy.cn,direct && go build -o quantumclaw . 2>&1 && echo BUILD_OK", { maxBuffer: 999999 }, (e2, stream) => {
        if(e2) { console.error(e2.message); conn.end(); return; }
        let out2 = "";
        stream.on("data", d => out2 += d.toString());
        stream.on("close", () => {
          console.log(out2);
          if(out2.includes("BUILD_OK")) {
            conn.exec("killall -9 quantumclaw 2>/dev/null; sleep 2; cd /opt/quantumclaw && nohup ./quantumclaw > quantumclaw.log 2>&1 & echo RESTARTED", (e3, s3) => {
              let o = "";
              s3.on("data", d => o += d.toString());
              s3.on("close", () => { console.log(o); conn.end(); });
            });
          } else conn.end();
        });
      });
    });
  });
});
conn.on("error", e => { console.error(e.message); process.exit(1); });
conn.connect({ host:"122.51.221.43", port:22, username:"ubuntu", password:"Ctji@2020", readyTimeout:600000 });
