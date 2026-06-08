#!/usr/bin/env python3
"""Deploy QuantumClaw to remote server via SSH.

Set the following environment variables before running:
  DEPLOY_HOST   - server IP or hostname (default: 127.0.0.1)
  DEPLOY_PORT   - SSH port (default: 22)
  DEPLOY_USER   - SSH username (default: ubuntu)
  DEPLOY_PASS   - SSH password (or use SSH key)
  DEPLOY_BIN    - local path to quantumclaw-linux binary
  DEPLOY_DIST   - local path to web/default/dist directory
"""

import paramiko
import os
import time
import sys

HOST = os.getenv("DEPLOY_HOST", "127.0.0.1")
PORT = int(os.getenv("DEPLOY_PORT", "22"))
USER = os.getenv("DEPLOY_USER", "ubuntu")
PASS = os.getenv("DEPLOY_PASS") or os.getenv("SSH_PASSWORD")
if not PASS:
    print("[ERROR] DEPLOY_PASS (or SSH_PASSWORD) environment variable is required.")
    print("  Example: export DEPLOY_PASS='your_password'")
    sys.exit(1)

LOCAL_PROJECT = os.getenv("DEPLOY_LOCAL_PROJECT", r"H:\AiData\openclaw\workspace\QuantumClaw")
LOCAL_BIN = os.getenv("DEPLOY_BIN", os.path.join(LOCAL_PROJECT, "quantumclaw-linux"))
LOCAL_DIST = os.getenv("DEPLOY_DIST", os.path.join(LOCAL_PROJECT, "web", "default", "dist"))

REMOTE_BIN_PATH = "/opt/quantumclaw/quantumclaw"
REMOTE_DIST = "/opt/quantumclaw/dist"
# Upload to ubuntu's home first, then sudo mv
HOME_BIN = "/home/ubuntu/quantumclaw_new"
HOME_DIST = "/home/ubuntu/qc_dist_new"

def ssh_connect():
    client = paramiko.SSHClient()
    client.set_missing_host_key_policy(paramiko.AutoAddPolicy())
    print(f"[INFO] Connecting to {USER}@{HOST}:{PORT} ...")
    client.connect(HOST, port=PORT, username=USER, password=PASS, timeout=15)
    print("[INFO] SSH connected!")
    return client

def run_cmd(client, cmd, timeout=120):
    stdin, stdout, stderr = client.exec_command(cmd, timeout=timeout)
    out = stdout.read().decode("utf-8", errors="replace")
    err = stderr.read().decode("utf-8", errors="replace")
    rc = stdout.channel.recv_exit_status()
    if out.strip():
        for line in out.strip().splitlines():
            print(f"[OUT] {line}")
    if err.strip() and rc != 0:
        for line in err.strip().splitlines():
            print(f"[ERR] {line}")
    return rc, out, err

def sftp_upload_dir(sftp, local_dir, remote_dir):
    try:
        sftp.listdir(remote_dir)
    except Exception:
        try:
            sftp.mkdir(remote_dir)
            print(f"[SFTP] Created dir: {remote_dir}")
        except Exception as e:
            print(f"[WARN] Cannot create {remote_dir}: {e}")
            return
    for item in os.listdir(local_dir):
        local_path = os.path.join(local_dir, item)
        remote_path = remote_dir + "/" + item
        if os.path.isdir(local_path):
            sftp_upload_dir(sftp, local_path, remote_path)
        else:
            sftp.put(local_path, remote_path)
            print(f"[SFTP] {item}")

def main():
    client = ssh_connect()
    sftp = client.open_sftp()

    # Step 1: Stop service
    print("\n=== Step 1: Stop Service ===")
    rc, _, _ = run_cmd(client, "sudo systemctl stop quantumclaw 2>&1", timeout=30)
    if rc != 0:
        print("[WARN] sudo stop failed, trying pkill...")
        run_cmd(client, "pkill -f quantumclaw || true", timeout=10)
    time.sleep(2)
    run_cmd(client, "systemctl status quantumclaw --no-pager 2>&1 | head -5 || echo 'Service stopped'")

    # Step 2: Upload binary to home dir
    print("\n=== Step 2: Upload Binary ===")
    if not os.path.isfile(LOCAL_BIN):
        print(f"[ERROR] Local binary not found: {LOCAL_BIN}")
        client.close()
        return

    print(f"[SFTP] Uploading {LOCAL_BIN} -> {HOME_BIN} ...")
    sftp.put(LOCAL_BIN, HOME_BIN)
    print("[SFTP] Upload done.")
    run_cmd(client, f"ls -lh {HOME_BIN}")

    # Step 3: Move binary to /opt/quantumclaw/ with sudo
    print(f"\n=== Step 3: Install Binary ===")
    run_cmd(client, f"sudo cp {HOME_BIN} {REMOTE_BIN_PATH} && sudo chmod +x {REMOTE_BIN_PATH}", timeout=30)
    run_cmd(client, f"ls -lh {REMOTE_BIN_PATH}")

    # Step 4: Upload frontend dist/ to home dir
    print("\n=== Step 4: Upload Frontend Dist ===")
    if os.path.isdir(LOCAL_DIST):
        print(f"[SFTP] Uploading dist/ to {HOME_DIST} ...")
        try:
            sftp.mkdir(HOME_DIST)
        except Exception:
            pass
        sftp_upload_dir(sftp, LOCAL_DIST, HOME_DIST)
        print("[SFTP] Upload done.")
        # Move to /opt/quantumclaw/dist with sudo
        run_cmd(client, f"sudo rm -rf {REMOTE_DIST}.bak 2>/dev/null; sudo mv {REMOTE_DIST} {REMOTE_DIST}.bak 2>/dev/null; sudo mv {HOME_DIST} {REMOTE_DIST}", timeout=60)
        run_cmd(client, f"ls -la {REMOTE_DIST}/ | head -10")
    else:
        print(f"[WARN] Local dist/ not found: {LOCAL_DIST}")

    # Step 5: Start service
    print("\n=== Step 5: Start Service ===")
    rc, _, _ = run_cmd(client, "sudo systemctl start quantumclaw 2>&1", timeout=30)
    if rc != 0:
        print("[WARN] sudo start failed, check service status...")
        run_cmd(client, "sudo systemctl status quantumclaw --no-pager 2>&1 | head -20")

    # Step 6: Verify
    print("\n=== Step 6: Verify ===")
    time.sleep(3)
    run_cmd(client, "systemctl status quantumclaw --no-pager 2>&1 | head -15")
    run_cmd(client, "ss -tlnp | grep 3666 || echo 'Port 3666 check done'")

    # Step 7: Health check
    print("\n=== Step 7: Health Check ===")
    time.sleep(2)
    run_cmd(client, "curl -s http://localhost:3666/api/status 2>/dev/null | head -20 || echo 'Health check done'")

    # Cleanup
    print("\n=== Cleanup ===")
    run_cmd(client, f"rm -f {HOME_BIN} && echo 'Cleaned up home dir'")

    sftp.close()
    client.close()
    print("\n✅ Deployment completed!")

if __name__ == "__main__":
    main()
