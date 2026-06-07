-- 000001: 初始 Schema（从原有 SQL/ 目录整合）
-- 包含：settlement_config, token_transaction, reseller,
--       affiliate_relation, platform_config, channel/ability 变更
-- 原文件：SQL/20260523_settlement_system.sql

-- ── 1. settlement_config ──────────────────────────────────────
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

INSERT INTO settlement_config (model_name, unified_cost, commission_rate, platform_fee_rate, enabled, created_time, updated_time)
VALUES ('*', 0.001000, 0.2000, 0.1000, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

-- ── 2. token_transaction ─────────────────────────────────────
CREATE TABLE IF NOT EXISTS token_transaction (
  id                  BIGINT PRIMARY KEY AUTO_INCREMENT,
  log_id              INT NOT NULL DEFAULT 0,
  user_id             INT NOT NULL DEFAULT 0,
  model_name          VARCHAR(255) NOT NULL DEFAULT '',
  prompt_tokens       INT NOT NULL DEFAULT 0,
  completion_tokens   INT NOT NULL DEFAULT 0,
  channel_id          INT NOT NULL DEFAULT 0,
  channel_owner_id    INT NOT NULL DEFAULT 0,
  promoter_id         INT NOT NULL DEFAULT 0,
  is_fallback         TINYINT(1) NOT NULL DEFAULT 0,
  unit_price          DECIMAL(16,8) NOT NULL DEFAULT 0,
  total_amount        DECIMAL(16,8) NOT NULL DEFAULT 0,
  unified_cost        DECIMAL(16,8) NOT NULL DEFAULT 0,
  commission_amount   DECIMAL(16,8) NOT NULL DEFAULT 0,
  platform_fee        DECIMAL(16,8) NOT NULL DEFAULT 0,
  key_provider_cost   DECIMAL(16,8) DEFAULT 0,
  created_time        BIGINT NOT NULL,
  INDEX idx_promoter (promoter_id),
  INDEX idx_channel_owner (channel_owner_id),
  INDEX idx_user (user_id),
  INDEX idx_model (model_name),
  INDEX idx_created (created_time)
);

-- ── 3. reseller ───────────────────────────────────────────────
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

-- ── 4. affiliate_relation ─────────────────────────────────────
CREATE TABLE IF NOT EXISTS affiliate_relation (
  id              INT PRIMARY KEY AUTO_INCREMENT,
  promoter_id     INT NOT NULL,
  consumer_id     INT NOT NULL,
  created_time    BIGINT NOT NULL,
  expires_at      BIGINT DEFAULT NULL,
  UNIQUE KEY idx_consumer (consumer_id)
);

-- ── 5. platform_config ────────────────────────────────────────
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

-- ── 6. channel 表变更 ────────────────────────────────────────
ALTER TABLE channel ADD COLUMN IF NOT EXISTS cost_price DECIMAL(10,6) DEFAULT 0 AFTER sell_price_rate;

-- ── 7. ability 表新增 user_id ────────────────────────────────
ALTER TABLE ability ADD COLUMN IF NOT EXISTS user_id INT DEFAULT 0 AFTER channel_id;
CREATE INDEX IF NOT EXISTS idx_ability_user ON ability(user_id);
