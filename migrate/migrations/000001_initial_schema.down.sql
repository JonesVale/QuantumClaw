-- 000001 回滚：删除结算系统相关表和字段

DROP TABLE IF EXISTS token_transaction;
DROP TABLE IF EXISTS reseller;
DROP TABLE IF EXISTS affiliate_relation;
DROP TABLE IF EXISTS settlement_config;
DROP TABLE IF EXISTS platform_config;

ALTER TABLE channel DROP COLUMN IF EXISTS cost_price;
ALTER TABLE ability DROP COLUMN IF EXISTS user_id;
