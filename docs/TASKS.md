# QuantumClaw 市场架构 — 完整任务清单

> 依托于 `docs/complete-architecture-v1.md` 的全链路分析
> 启动: 2026-06-06 | 预估: 4 周

---

## 依赖关系总图

```
Phase 1 ────────────────────→ Phase 2 ────────────────→ Phase 3 ──────────→ Phase 4
  店铺体系                      市场与上架                 结算与路由           评价与运营

1.1 Store 模型 ──────────→ 2.1 Listing 模型 ────────→ 3.5 市场路由
     │                            │                        │
     ├──→ 1.6 开店 API           ├──→ 2.2 上架 API        └──→ 3.6 平台池
     │                            │
     └──→ 1.9 store_service       ├──→ 2.3 市场行情 API
                                  │
1.2 FeeConfig 模型                ├──→ 2.5 Listing→Channel
     │                            │
     ├──→ 1.7 Admin 配置 API     └──→ 2.6 用户偏好
     │
     ├──→ 1.8 getStoreFeeRate ──→ 3.1 阶梯入驻费 ───→ 3.4 cron扩展
     │                                                │
1.3 Review 模型 ───────────────────────────────────────└──→ 4.1 评价 API
                                                              │
1.4 Channel 加字段                                              └──→ 4.2 信誉分
     │
     └──→ 3.6 平台池
```

---

## Phase 1: 店铺体系（6 月 7 日 - 6 月 13 日）

### P1.1 数据模型
- [ ] `model/store.go` — Store / StoreTierLog / StoreStatus 常量
- [ ] `model/fee_config.go` — PlatformFeeConfig + CRUD
- [ ] `model/review.go` — Review 结构体
- [ ] `model/channel.go` — 加 StoreID + IsPlatformPool 字段
- [ ] `model/main.go` — AutoMigrate 新表 + Seed 默认数据
- **文件**: 4 个新文件 + 2 个改文件 | **行数**: ~200

### P1.2 API 层
- [ ] `controller/store.go` — POST /api/store/register（开店）
- [ ] `controller/store.go` — GET /api/store/profile（店铺信息）
- [ ] `controller/store.go` — PUT /api/store/profile（修改信息）
- [ ] `controller/admin_store.go` — GET/PUT store-tiers（入驻费配置）
- **文件**: 2 个新文件 | **行数**: ~270

### P1.3 业务逻辑
- [ ] `service/store_service.go` — 开店校验 + 等级变更记录
- [ ] `model/platform_fee.go` — getPlatformFeeRate → getStoreFeeRate
- **文件**: 1 个新文件 + 1 个改文件 | **行数**: ~130

### P1.4 前端
- [ ] 开店表单页面
- [ ] 管理员入驻费配置页面

---

## Phase 2: 市场与上架（6 月 14 日 - 6 月 20 日）

### P2.1 数据模型
- [ ] `model/store.go` — Listing 模型补充（PricePerUnit/Region/Status/Stats）
- [ ] `model/store.go` — Listing CRUD 函数
- **文件**: 1 个改文件 | **行数**: ~60

### P2.2 API 层
- [ ] `controller/store.go` — POST/GET/PUT listings（上架/列表/调价/下架）
- [ ] `controller/market.go` — GET /market/models（模型列表+最低价）
- [ ] `controller/market.go` — GET /market/price/:model（各店铺报价）
- [ ] `controller/market.go` — GET /market/stores（店铺列表+搜索）
- [ ] `controller/market.go` — GET /market/listing/:id（上架详情）
- [ ] `controller/user.go` — PUT/GET preferred_store（用户偏好）
- **文件**: 2 个新文件 + 1 个改文件 | **行数**: ~390

### P2.3 业务逻辑
- [ ] `service/market_service.go` — 搜索/比价/排序
- [ ] `service/store_service.go` — Listing→Channel 自动关联
- [ ] `controller/channel-test.go` — Listing 接入健康检测
- **文件**: 1 个新文件 + 2 个改文件 | **行数**: ~160

### P2.4 前端
- [ ] 市场行情页（模型列表 + 价格 + 店铺）
- [ ] Provider 上架资源表单
- [ ] 用户偏好店铺设置

---

## Phase 3: 结算与路由（6 月 21 日 - 6 月 27 日）

### P3.1 入驻费阶梯
- [ ] `model/platform_fee.go` — AutoSettleMonthlyFees 按 Store 等级遍历
- [ ] `service/store_service.go` — autoUpgradeStoreTier（自动升降级）
- [ ] `model/commission.go` — 提现关联 StoreID
- [ ] `main.go` — cron 扩展
- **文件**: 3 个改文件 | **行数**: ~120

### P3.2 路由改造
- [ ] `middleware/distributor.go` — 三层 spill-over（第三方最低价→其他→平台池）
- [ ] `model/channel.go` — IsPlatformPool 路由支持
- [ ] `controller/channel-admin.go` — 平台池管理 API
- **文件**: 2 个改文件 + 0 个新文件 | **行数**: ~110

### P3.3 前端
- [ ] Provider 收益看板（含入驻费明细）
- [ ] 平台池管理页面

---

## Phase 4: 评价与运营（6 月 28 日 - 7 月 4 日）

### P4.1 评价
- [ ] `controller/market.go` — POST /market/reviews（提交评价）
- [ ] `controller/market.go` — GET /market/reviews（查看评价）
- **文件**: 1 个改文件 | **行数**: ~60

### P4.2 信誉分
- [ ] `service/market_service.go` — Store.Rating 自动计算
- [ ] `service/market_service.go` — 搜索排序接入信誉分
- **文件**: 1 个改文件 | **行数**: ~70

### P4.3 运营工具
- [ ] `controller/admin.go` — GET /admin/metrics（核心运营指标）
- [ ] `controller/admin_store.go` — 店铺详情+暂停/关店
- [ ] `model/store.go` — 数据统计查询
- **文件**: 2 个改文件 | **行数**: ~120

### P4.4 前端
- [ ] 运营仪表盘
- [ ] 管理员店铺管理页面（列表/搜索/调等级/暂停）
- [ ] 用户端店铺详情页

---

## 分文件新增/改动总结

### 新增文件

| 文件 | 归属 Phase | 预估行数 |
|------|-----------|---------|
| `model/store.go` | P1 + P2 | 140 |
| `model/fee_config.go` | P1 | 40 |
| `model/review.go` | P1 | 30 |
| `controller/store.go` | P1 + P2 | 350 |
| `controller/market.go` | P2 + P4 | 210 |
| `controller/admin_store.go` | P1 + P4 | 160 |
| `service/store_service.go` | P1 + P3 | 150 |
| `service/market_service.go` | P2 + P4 | 130 |
| **新增总计** | **8 个文件** | **~1210** |

### 改动文件

| 文件 | 归属 Phase | 预估改动行 |
|------|-----------|----------|
| `model/channel.go` | P1 | 20 |
| `model/platform_fee.go` | P1 + P3 | 90 |
| `model/commission.go` | P3 | 30 |
| `model/main.go` | P1 | 30 |
| `main.go` | P3 | 30 |
| `middleware/distributor.go` | P3 | 80 |
| `controller/user.go` | P2 | 40 |
| `controller/channel-test.go` | P2 | 40 |
| `controller/admin.go` | P4 | 80 |
| **改动总计** | **9 个文件** | **~440** |

### 最终统计

| 指标 | 数量 |
|------|------|
| 新增文件 | 8 个 |
| 改动文件 | 9 个 |
| 不改动文件 | ~60 个（现有代码的 75% 完全不动） |
| 新增行数 | ~1210 |
| 改动行数 | ~440 |
| **总变更** | **~1650 行** |

---

## 里程碑验收标准

| 里程牌 | 时间 | 验证方式 |
|--------|------|---------|
| 🏁 M1: 能开店 | W1 结束 | 任注册用户 POST /api/store/register → 返回 Store，入驻费默认 basic 10% |
| 🏁 M2: 能上架能逛 | W2 结束 | POST /api/store/listings → 上架成功；GET /api/market/models → 看到价格 |
| 🏁 M3: 钱能走对 | W3 结束 | 每月 1 号入驻费按阶梯算，提现自动扣除；路由优先选最低价 Store |
| 🏁 M4: 市场能自转 | W4 结束 | 评价生效，信誉分影响排序，运营面板看到数据 |

---

## 不变的部分

以下模块完全不碰，保持现状：

```
relay/           ← 适配器/中继/计费钩子
common/encrypt/  ← 加解密
common/config/   ← 配置
middleware/ 除了 distributor.go ← 认证/限流/安全
controller/auth/ ← OAuth 登录
controller/topup_* ← 支付
model/ 除了上面列出的 ← 大部分
router/          ← 路由注册（新增文件在 router/ 不动）
docs/            ← 文档
```

## 开始

从 `model/store.go` 的 Store 结构体开始写。这是所有后续代码的根基。
