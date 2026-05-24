-- QuantumClaw test seed data v2 (corrected table schemas)

-- 1. Insert simple logs (manually, since log IDs are auto)
-- Already have 4 logs from system startup, let me add user activity logs
INSERT IGNORE INTO logs (user_id, type, content, created_at) VALUES
(6, 1, '用户登录系统', UNIX_TIMESTAMP(NOW() - INTERVAL 2 HOUR) * 1000),
(6, 2, '调用 GPT-4 API: success', UNIX_TIMESTAMP(NOW() - INTERVAL 1 HOUR) * 1000),
(6, 2, '调用 Claude-3 Opus: success', UNIX_TIMESTAMP(NOW() - INTERVAL 30 MINUTE) * 1000),
(7, 1, '新用户注册（通过推荐链接 admin）', UNIX_TIMESTAMP(NOW()) * 1000);

-- 2. Platform config (key-value store)
INSERT IGNORE INTO platform_config (`key`, `value`) VALUES
('promotion_duration_days', '30'),
('settlement_cycle_days', '7'),
('min_withdrawal', '10'),
('commission_rate', '20'),
('referral_bonus', '50000');

-- 3. Settlement config (model pricing)
DELETE FROM settlement_config;
INSERT INTO settlement_config (model_name, unified_cost, commission_rate, platform_fee_rate) VALUES
('gpt-4', 10.0, 0.20, 0.05),
('gpt-4-turbo', 5.0, 0.20, 0.05),
('gpt-3.5-turbo', 1.0, 0.15, 0.05),
('claude-3-opus', 15.0, 0.20, 0.05),
('claude-3-sonnet', 3.0, 0.15, 0.05),
('deepseek-chat', 0.5, 0.10, 0.03),
('gemini-pro', 1.0, 0.10, 0.03);

-- 4. Commission settings (global)
-- Note: commission_settings table has different schema (global, not per-user)
-- Only need to ensure defaults exist
INSERT IGNORE INTO commission_settings (id, enabled, register_reward, consume_rate, min_withdraw) VALUES
(1, 1, 50000, 0.10, 10000);

-- 5. Token transactions (simulate team member usage)
INSERT IGNORE INTO token_transaction (user_id, promoter_id, channel_owner_id, model_name, prompt_tokens, completion_tokens, total_amount, commission_amount, platform_fee, unified_cost, created_time) VALUES
(7, 6, 0, 'gpt-4-mini', 500, 1000, 0.0015, 0.0003, 0.000075, 0.0005, UNIX_TIMESTAMP(NOW() - INTERVAL 1 HOUR) * 1000),
(7, 6, 0, 'gpt-4-mini', 800, 1700, 0.0025, 0.0005, 0.000125, 0.0005, UNIX_TIMESTAMP(NOW() - INTERVAL 30 MINUTE) * 1000);
