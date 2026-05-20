# QuantumClaw 全面完善任务规划

## 当前项目状态

```
健康度: build ✅ | vet ✅ | tests 12/12 ✅ | TODOs 0
已实现: 量子枚举 → 适配器 → 路由 → 控制器 → 计费 → 动态类型 → 开发者UX
已修复: 渠道测试 → 批量测试 → 重试 → i18n 补全 → API Docs
未推送: 15 commits 本地
```

---

## 一、核心决策：语言体系统一（P0）

### 现状 vs 目标

| 维度 | 当前 (react-i18next JSON) | 目标 (T_Languages 数据库) |
|------|--------------------------|--------------------------|
| 翻译存储 | 7 个 JSON 文件，~210 键/文件 | MySQL T_Languages 表 |
| 加载方式 | 编译时嵌入 | 运行时 API 拉取 |
| 切换 | `i18n.changeLanguage('zh-CN')` | `POST /api/language/switch { "lang": "中文简体" }` |
| 管理 | 改 JSON → 重新构建 → 部署 | 后台管理页面 → 保存即生效 |
| LanguagesType | `'zh-CN'` `'en'` (locale 代码) | `'中文简体'` `'English'` (显示名=查询条件) |

### 迁移方案（MySQL 本地，无 MSSQL）

```
                          ┌──────────────────┐
                          │   MySQL DB       │
                          │  T_Languages     │
                          │  LanguageTypes   │
                          └────────┬─────────┘
                                   │
                          ┌────────┴─────────┐
                          │  Backend API     │
                          │  GET /api/languages    │
                          │  GET /api/translations │
                          │  POST /api/language/switch │
                          └────────┬─────────┘
                                   │
                          ┌────────┴─────────┐
                          │  Frontend Hook   │
                          │  useTranslation()│
                          │  → fetch from API│
                          └─────────────────┘
```

### 迁移步骤

| # | 任务 | 文件 | 代码量 |
|---|------|------|--------|
| 1.1 | Go: 新增 LanguageTypes / T_Languages Model | `model/language.go` | ~60 行 |
| 1.2 | Go: 新增翻译 API (list / get / switch) | `controller/language.go` | ~80 行 |
| 1.3 | Go: 注册路由 | `router/api.go` | +2 行 |
| 1.4 | Go: 启动时从 DB 加载翻译到内存缓存 | `model/main.go` | +10 行 |
| 1.5 | Go: 初始化脚本插入种子翻译数据 | `model/language.go` | ~210 行（从 JSON 转 SQL） |
| 1.6 | 前端: 替换 react-i18next → 自定义 useTranslation | `src/hooks/useTranslation.ts` | ~80 行 |
| 1.7 | 前端: 重构所有 `t('key')` 调用 | 所有 tsx 文件 | 批量替换 |
| 1.8 | 前端: 语言切换改为 API 驱动 | 语言选择器组件 | ~20 行 |

**预估: 3~4 小时**

> ⚠️ 这个阶段会影响所有前端页面。做完之后后续所有 UI 改动都基于新语言体系。

---

## 二、UI 前端展示补全（P1）

> 必须在语言体系统一之后做，否则用了新体系又要重改翻译

### 2.1 Landing 首页 + 量子算力

| 改动 | 说明 | 预估 |
|------|------|------|
| Features 区加新卡片「Quantum Computing」 | 量子算力分发、多平台支持 | 30 分钟 |
| Hero 区加「Quantum」tag | 与 AI 模型并列展示 | 15 分钟 |
| i18n 补 4 个翻译键 | 标题+描述（数据库方式） | 10 分钟 |

### 2.2 Models 页 + 量子后端列表

| 改动 | 说明 | 预估 |
|------|------|------|
| 后端: `ListAllModels` 返回 quantum backends | 从 Ability 表或 Channel 表取 | 20 分钟 |
| 前端: Models 页区分 AI/Quantum tab | 类似 Channels 页的 tab 模式 | 30 分钟 |

### 2.3 Billing 页 + 量子消耗拆分

| 改动 | 说明 | 预估 |
|------|------|------|
| 前端: 区分 AI/量子配额消耗 | 从 Log 表的 ModelName 区分 | 20 分钟 |

### 2.4 Monitoring 页 + AI/量子分类

| 改动 | 说明 | 预估 |
|------|------|------|
| 前端: 加 type tab 筛选 | 类似 Channels 页分类方式 | 20 分钟 |

**预估: 2~3 小时**

---

## 三、后台管理补全（P2）

### 3.1 Settings → Billing → 量子计费模板

| 改动 | 说明 | 预估 |
|------|------|------|
| 后端: 量子计费默认表达式 | 预置 `qubits * shots * 0.001` | 10 分钟 |
| 前端: Billing tab 加量子区域 | 展示模板 + 可编辑 | 40 分钟 |

### 3.2 Tasks 页 + 量子任务

| 改动 | 说明 | 预估 |
|------|------|------|
| 前端: 量子任务 tab | 从 Log 表显示量子任务记录 | 30 分钟 |

**预估: 1.5 小时**

---

## 四、文档 + 发布（P3）

| # | 任务 | 说明 | 预估 |
|---|------|------|------|
| 4.1 | README 补充量子算力章节 | 中/英/日三语 | 30 分钟 |
| 4.2 | 版本号 v2.1.0 → v2.2.0 | VERSION 文件 | 1 分钟 |
| 4.3 | 全量推送 remote | git push | 网络通即可 |

**预估: 30 分钟**

---

## 五、依赖关系和执行顺序

```
语言体系统一 (3-4h) ←─ 必须先做，后面所有 UI 基于新体系
  │
  ├── Landing 首页+量子 (30min)
  ├── Models 页+量子后端 (50min)
  ├── Billing 页+拆分 (20min)
  └── Monitoring 分类 (20min)
        │
        ├── Settings 计费模板 (50min)
        └── Tasks 量子任务 (30min)
              │
              └── README + 版本 + 推送 (30min)
```

**关键依赖线：**
1. 语言体系 → 所有 UI 改动
2. Landing → 首页展示
3. Models → 后端先改 `ListAllModels`
4. Monitoring → 无外部依赖
5. Settings → 无外部依赖
6. README → 最后做

---

## 六、要不要做语言体系统一？

### 做的好处
- 和官网（CTJIWeb）架构统一，维护一处翻译数据
- 管理员可以在后台直接改翻译，不需重新构建前端
- 翻译数据随数据库备份，不丢失

### 不做的理由
- 当前 react-i18next JSON 方式已正常运行，全球化没有问题
- 迁移成本 ~3~4 小时，涉及所有前端页面
- QuantumClaw 和 CTJIWeb 本质是两个独立项目，语言体系分开不影响各自运行

### 我的建议

**先推已完成的工作，语言体系统一独立评估。**

当前 15 个 commit 全部是量子算力 + 质量增强，与语言体系无关。先把这些推 remote，避免本地丢失。然后你决定要不要做语言统一，再做 UI 补全。

---

## 七、如果推 remote — 执行清单

```bash
git push origin master
# 推送 15 个 commit (8401b36 → c1cd17e)
```

之后我们讨论：
- 要不要做语言统一
- 还是按当前 react-i18next 继续做 UI 补全
