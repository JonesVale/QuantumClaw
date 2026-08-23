# QuantumClaw 版权信息更新

## 正确的版权信息

**版权所有 (C) 2026 深圳市中科劲纬智能有限公司**  
**官网**: https://www.ctji.cn  
**邮箱**: vale@ctji.cn

---

## 需要更新的文件清单

请手动更新以下文件中的邮箱地址（从 business@ctji.cn 改为 vale@ctji.cn）：

### 1. README.md
```markdown
**邮箱**: vale@ctji.cn
```

### 2. LICENSE
```markdown
Email: vale@ctji.cn
```

### 3. LICENSE_COMMERCIAL.md
```markdown
- **邮箱**: vale@ctji.cn
```

### 4. CONTRIBUTING.md
```markdown
- **邮箱**: vale@ctji.cn
```

### 5. CHANGELOG.md
```markdown
- **邮箱**: vale@ctji.cn
```

### 6. docs/DEPLOYMENT.md
```markdown
- **邮箱**: vale@ctji.cn
```

### 7. PROJECT_STRUCTURE.md
```markdown
*邮箱: vale@ctji.cn*
```

### 8. SECURITY.md
```markdown
- **邮箱**: vale@ctji.cn
```

---

## 批量更新命令（Linux/Mac）

```bash
cd QuantumClaw-Open
find . -name "*.md" -exec sed -i 's/business@ctji.cn/vale@ctji.cn/g' {} \;
find . -name "*.md" -exec sed -i 's/support@ctji.cn/vale@ctji.cn/g' {} \;
```

## 批量更新命令（Windows PowerShell）

```powershell
cd QuantumClaw-Open
Get-ChildItem -Recurse -Filter "*.md" | ForEach-Object {
    (Get-Content $_.FullName) -replace 'business@ctji.cn', 'vale@ctji.cn' -replace 'support@ctji.cn', 'vale@ctji.cn' | Set-Content $_.FullName
}
```

---

## 已确认的版权信息

| 项目 | 内容 |
|------|------|
| 公司全称 | 深圳市中科劲纬智能有限公司 |
| 官网 | https://www.ctji.cn |
| 联系邮箱 | vale@ctji.cn |
| 开源许可证 | AGPL-3.0 |
| 商业许可证 | 需单独购买 |

---

*更新时间：2026-08-23*
