# QuantumClaw 分支策略

## 分支结构

```
master          ← 生产分支，仅通过 PR 合并
  ↑
develop         ← 集成分支，功能分支在此合并验证
  ↑
feature/*       ← 功能分支，从 develop 拉出
fix/*           ← 修复分支
hotfix/*        ← 紧急修复，从 master 拉出，合并回 master + develop
```

## 工作流

### 日常开发
```bash
git checkout develop
git pull origin develop
git checkout -b feature/my-feature
# ... 开发 ...
git add .
git commit -m "feat: 描述"
git checkout develop
git merge feature/my-feature
git branch -d feature/my-feature   # 合并后删除
```

### 发布到生产
```bash
# 在 develop 上验证通过后
git checkout master
git merge develop
git tag v2.3.0
git push origin master --tags
```

### 紧急修复
```bash
git checkout master
git checkout -b hotfix/critical-bug
# ... 修复 ...
git checkout master
git merge hotfix/critical-bug
git checkout develop
git merge hotfix/critical-bug      # 修复必须同步到 develop
git branch -d hotfix/critical-bug
```

## 分支命名规范

| 分支类型 | 格式 | 例子 |
|---------|------|------|
| 功能开发 | `feature/<简短描述>` | `feature/alipay-direct` |
| 修复 | `fix/<问题描述>` | `fix/i18n-missing-keys` |
| 紧急修复 | `hotfix/<严重问题>` | `hotfix/payment-signature` |
| 重构 | `refactor/<模块名>` | `refactor/billing-engine` |
| 文档 | `docs/<内容>` | `docs/api-guide` |

## 提交信息规范

```
<type>: <简短描述>

<可选详细说明>
```

类型: `feat` `fix` `chore` `refactor` `docs` `style` `test`

## 保护规则（GitHub 启用后）

- master: 禁止直接推送，仅允许 PR 合并
- develop: 建议保护，允许直接推送
- PR 合并前需通过 CI（build + vet + test）
