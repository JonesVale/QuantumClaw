-- ============================================================================
-- QuantumClaw 结算与路由系统 - 数据库迁移
-- Phase 1: 路由 failover + 结算流水
-- Phase 2: 推广关系 + 渠道商管理
-- ============================================================================
-- 注意：所有用户可见文本不在此处，在 T_Languages 表中

-- ── 1. settlement_config（结算配置：每个模型一条）────────────────────────
CREATE TABLE IF NOT EXISTS settlement_config (
  id                INT PRIMARY KEY AUTO_INCREMENT,
  model_name        VARCHAR(255) NOT NULL,
  unified_cost      DECIMAL(10,6) NOT NULL DEFAULT 0.001000,
  commission_rate   DECIMAL(5,4) NOT NULL DEFAULT 0.2000,
  platform_fee_rate DECIMAL(5,4) NOT NULL DEFAULT 0.1000,
  enabled           TINYINT(1) DEFAULT 1,
  created_time      BIGINT NOT NULL,
  updated_time      BIGINT NOT NULL,
  UNIQUE KEY idx_model_name (model_name)
);

-- 默认配置行
INSERT INTO settlement_config (model_name, unified_cost, commission_rate, platform_fee_rate, enabled, created_time, updated_time)
VALUES ('*', 0.001000, 0.2000, 0.1000, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

-- ── 2. token_transaction（交易流水）─────────────────────────────────────
CREATE TABLE IF NOT EXISTS token_transaction (
  id                  BIGINT PRIMARY KEY AUTO_INCREMENT,

  -- 调用基本信息
  log_id              INT NOT NULL DEFAULT 0,
  user_id             INT NOT NULL DEFAULT 0,
  model_name          VARCHAR(255) NOT NULL DEFAULT '',
  prompt_tokens       INT NOT NULL DEFAULT 0,
  completion_tokens   INT NOT NULL DEFAULT 0,

  -- 路由信息
  channel_id          INT NOT NULL DEFAULT 0,
  channel_owner_id    INT NOT NULL DEFAULT 0,
  promoter_id         INT NOT NULL DEFAULT 0,
  is_fallback         TINYINT(1) NOT NULL DEFAULT 0,

  -- 金额字段（USD）
  unit_price          DECIMAL(16,8) NOT NULL DEFAULT 0,
  total_amount        DECIMAL(16,8) NOT NULL DEFAULT 0,
  unified_cost        DECIMAL(16,8) NOT NULL DEFAULT 0,
  commission_amount   DECIMAL(16,8) NOT NULL DEFAULT 0,
  platform_fee        DECIMAL(16,8) NOT NULL DEFAULT 0,

  -- Key 贡献者成本（由 Key 主填写，仅对账使用）
  key_provider_cost   DECIMAL(16,8) DEFAULT 0,

  created_time        BIGINT NOT NULL,

  INDEX idx_promoter (promoter_id),
  INDEX idx_channel_owner (channel_owner_id),
  INDEX idx_user (user_id),
  INDEX idx_model (model_name),
  INDEX idx_created (created_time)
);

-- ── 3. reseller（渠道商/推广者）──────────────────────────────────────────
CREATE TABLE IF NOT EXISTS reseller (
  id                INT PRIMARY KEY AUTO_INCREMENT,
  user_id           INT NOT NULL,
  name              VARCHAR(255) NOT NULL DEFAULT '',
  description       TEXT,
  contact_info      TEXT,
  avatar_url        VARCHAR(1024) DEFAULT '',
  status            TINYINT(1) DEFAULT 1,
  affiliate_code    VARCHAR(64) UNIQUE,
  balance           DECIMAL(16,8) DEFAULT 0,
  total_earned      DECIMAL(16,8) DEFAULT 0,
  created_time      BIGINT NOT NULL,
  updated_time      BIGINT NOT NULL,
  UNIQUE KEY idx_user (user_id)
);

-- ── 4. affiliate_relation（推广关系）─────────────────────────────────────
CREATE TABLE IF NOT EXISTS affiliate_relation (
  id              INT PRIMARY KEY AUTO_INCREMENT,
  promoter_id     INT NOT NULL,
  consumer_id     INT NOT NULL,
  created_time    BIGINT NOT NULL,
  expires_at      BIGINT DEFAULT NULL,
  UNIQUE KEY idx_consumer (consumer_id)
);

-- ── 5. platform_config（全局配置）────────────────────────────────────────
CREATE TABLE IF NOT EXISTS platform_config (
  `key`   VARCHAR(255) PRIMARY KEY,
  `value` TEXT NOT NULL,
  updated_time BIGINT NOT NULL
);

INSERT INTO platform_config (`key`, `value`, updated_time) VALUES
  ('promotion_expire_days', '0', UNIX_TIMESTAMP()),
  ('settlement_cycle', 'monthly', UNIX_TIMESTAMP()),
  ('min_withdraw_amount', '10', UNIX_TIMESTAMP())
ON DUPLICATE KEY UPDATE `value` = VALUES(`value`), updated_time = UNIX_TIMESTAMP();

-- ── 6. channel 表变更 ───────────────────────────────────────────────────
ALTER TABLE channel ADD COLUMN IF NOT EXISTS cost_price DECIMAL(10,6) DEFAULT 0 AFTER sell_price_rate;

-- ── 7. ability 表新增 user_id 字段 ──────────────────────────────────────
ALTER TABLE ability ADD COLUMN IF NOT EXISTS user_id INT DEFAULT 0 AFTER channel_id;
CREATE INDEX IF NOT EXISTS idx_ability_user ON ability(user_id);
