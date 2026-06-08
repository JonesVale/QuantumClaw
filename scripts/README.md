# QuantumClaw 脚本目录

本目录包含 QuantumClaw 的运维脚本。

---

## 备份与恢复

### `backup.sh` — 备份 QuantumClaw

**功能**：
- 备份数据库（SQLite 或 MySQL）
- 备份 `.env` 配置文件
- 备份 `.session_secret`（Session 密钥）
- 备份加密密钥（`CryptoSecret` / `SessionSecret` / `JWTSecret`）

**用法**：
```bash
# 备份到默认目录（./backups/）
bash scripts/backup.sh

# 备份到指定目录
bash scripts/backup.sh /path/to/backups
```

**备份内容**：
```
quantumclaw-backup-20260608-120000.tar.gz
├── quantumclaw.db          # SQLite 数据库（如使用 SQLite）
├── mysql_dump.sql          # MySQL dump（如使用 MySQL）
├── .env                    # 配置文件
├── .session_secret         # Session 密钥
└── keys.txt                # 加密密钥（可选）
```

---

### `restore.sh` — 恢复 QuantumClaw

**功能**：
- 恢复 `.env` 配置文件
- 恢复 `.session_secret`
- 恢复数据库（SQLite 直接复制；MySQL 提示手动导入）
- 恢复加密密钥

**用法**：
```bash
# 1. 解压备份
tar -xzf quantumclaw-backup-20260608-120000.tar.gz -C /tmp

# 2. 恢复
bash scripts/restore.sh /tmp/quantumclaw-backup-20260608-120000
```

**注意**：
- 恢复前会备份当前 `.env`（防止恢复失败）
- MySQL 恢复需要手动执行 `mysql` 导入命令
- 恢复后需重启 QuantumClaw

---

## 备份策略建议

| 场景 | 频率 | 保留份数 |
|------|------|---------|
| 开发环境 | 每天 | 7 份 |
| 生产环境 | 每天 | 30 份 |
| 关键操作前 | 立即 | 1 份 |

**自动备份（crontab）**：
```bash
# 每天凌晨 2 点备份
0 2 * * * cd /path/to/QuantumClaw && bash scripts/backup.sh /backups/quantumclaw

# 删除 30 天前的备份
0 3 * * * find /backups/quantumclaw -name "quantumclaw-backup-*.tar.gz" -mtime +30 -delete
```

---

## 故障排查

### 备份失败

**问题**：`mysqldump` 未找到  
**解决**：安装 MySQL 客户端
```bash
# Ubuntu/Debian
sudo apt install mysql-client

# CentOS/RHEL
sudo yum install mysql
```

### 恢复失败

**问题**：恢复后无法登录  
**原因**：`SESSION_SECRET` 或 `.session_secret` 不匹配  
**解决**：确保恢复后使用相同的 `SESSION_SECRET`

---

## 安全注意事项

⚠️ **备份文件包含敏感信息**：
- `.env` 包含数据库密码、API Key
- `CryptoSecret` 用于解密 Channel Key
- `SessionSecret` 用于 Session 签名

**建议**：
1. 备份文件权限设为 `600`（仅所有者可读）
2. 备份文件加密存储（GPG / AES）
3. 不要将备份文件提交到 Git
4. 定期检查备份完整性

---

*最后更新：2026-06-08*
