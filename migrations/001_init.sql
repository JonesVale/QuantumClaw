-- QuantumClaw Open Source Edition 数据库初始化脚本
-- 版权所有 (C) 2026 深圳市中科劲纬智能有限公司
-- 官网: https://www.ctji.cn
-- 许可证: AGPL-3.0

-- 创建数据库
CREATE DATABASE IF NOT EXISTS quantumclaw_open CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

USE quantumclaw_open;

-- 用户表
CREATE TABLE IF NOT EXISTS users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    username VARCHAR(50) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE,
    role INT DEFAULT 1 COMMENT '0=访客,1=普通用户,2=供应商,10=管理员,100=根用户,1000=分销商',
    status INT DEFAULT 1 COMMENT '1=启用,2=禁用,3=已注销',
    quota BIGINT DEFAULT 0 COMMENT '配额',
    used_quota BIGINT DEFAULT 0 COMMENT '已用配额',
    cash_balance BIGINT DEFAULT 0 COMMENT '现金余额(分)',
    commission_balance BIGINT DEFAULT 0 COMMENT '佣金余额(分)',
    debt BIGINT DEFAULT 0 COMMENT '欠费(分)',
    aff_code VARCHAR(32) UNIQUE,
    inviter_id INT DEFAULT 0,
    group VARCHAR(32) DEFAULT 'default',
    created_at BIGINT NOT NULL,
    updated_at BIGINT NOT NULL,
    INDEX idx_username (username),
    INDEX idx_email (email),
    INDEX idx_aff_code (aff_code),
    INDEX idx_inviter (inviter_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 渠道表
CREATE TABLE IF NOT EXISTS channels (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    type INT NOT NULL COMMENT '渠道类型',
    api_key VARCHAR(255) NOT NULL,
    base_url VARCHAR(255),
    cost_per_unit DECIMAL(10,6) DEFAULT 0.002,
    sell_price_rate DECIMAL(5,4) DEFAULT 1.5,
    profit_split DECIMAL(5,4) DEFAULT 0.8,
    status INT DEFAULT 1,
    created_at BIGINT NOT NULL,
    updated_at BIGINT NOT NULL,
    INDEX idx_type (type),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 订阅计划表
CREATE TABLE IF NOT EXISTS subscription_plans (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    price_cents INT NOT NULL,
    quota INT NOT NULL,
    duration_unit VARCHAR(20) DEFAULT 'month',
    duration_value INT DEFAULT 1,
    status INT DEFAULT 1,
    created_at BIGINT NOT NULL,
    updated_at BIGINT NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 用户订阅表
CREATE TABLE IF NOT EXISTS user_subscriptions (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL,
    plan_id INT NOT NULL,
    start_time BIGINT NOT NULL,
    end_time BIGINT NOT NULL,
    next_reset_time BIGINT DEFAULT 0,
    status VARCHAR(20) DEFAULT 'active',
    created_at BIGINT NOT NULL,
    updated_at BIGINT NOT NULL,
    INDEX idx_user (user_id),
    INDEX idx_plan (plan_id),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 分销商表
CREATE TABLE IF NOT EXISTS distributors (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL,
    name VARCHAR(100),
    contact_email VARCHAR(255),
    api_key VARCHAR(64) UNIQUE,
    markup_rate DECIMAL(5,4) DEFAULT 0,
    profit_split DECIMAL(5,4) DEFAULT 0.5,
    status INT DEFAULT 1,
    created_at BIGINT NOT NULL,
    updated_at BIGINT NOT NULL,
    INDEX idx_user (user_id),
    INDEX idx_api_key (api_key)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 级联节点表
CREATE TABLE IF NOT EXISTS cascade_nodes (
    id INT AUTO_INCREMENT PRIMARY KEY,
    node_key VARCHAR(64) UNIQUE NOT NULL,
    name VARCHAR(100),
    api_key VARCHAR(64) UNIQUE,
    status INT DEFAULT 1 COMMENT '1=在线,2=离线',
    last_heartbeat BIGINT DEFAULT 0,
    created_at BIGINT NOT NULL,
    updated_at BIGINT NOT NULL,
    INDEX idx_node_key (node_key),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 异步任务表
CREATE TABLE IF NOT EXISTS async_tasks (
    id INT AUTO_INCREMENT PRIMARY KEY,
    task_id VARCHAR(64) UNIQUE NOT NULL,
    user_id INT NOT NULL,
    task_type VARCHAR(50) NOT NULL COMMENT 'midjourney,video,suno',
    status VARCHAR(20) DEFAULT 'pending',
    quota INT DEFAULT 0,
    created_at BIGINT NOT NULL,
    updated_at BIGINT NOT NULL,
    finished_at BIGINT DEFAULT 0,
    INDEX idx_task_id (task_id),
    INDEX idx_user (user_id),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 配置表
CREATE TABLE IF NOT EXISTS system_config (
    id INT AUTO_INCREMENT PRIMARY KEY,
    `key` VARCHAR(100) UNIQUE NOT NULL,
    `value` TEXT,
    description VARCHAR(255),
    created_at BIGINT NOT NULL,
    updated_at BIGINT NOT NULL,
    INDEX idx_key (`key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 插入默认配置
INSERT INTO system_config (`key`, `value`, `description`, created_at, updated_at) VALUES
('db_version', '1.0.0', '数据库版本', UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
('site_name', 'QuantumClaw Open', '网站名称', UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
('jwt_secret', 'change_me_in_production', 'JWT密钥', UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

-- 插入默认管理员用户 (密码: admin123)
INSERT INTO users (username, password, email, role, status, quota, created_at, updated_at) VALUES
('admin', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 'admin@ctji.cn', 10, 1, 999999999, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());

-- 插入默认订阅计划
INSERT INTO subscription_plans (name, description, price_cents, quota, duration_unit, duration_value, status, created_at, updated_at) VALUES
('免费版', '基础功能', 0, 1000, 'month', 1, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
('标准版', '更多配额', 999, 10000, 'month', 1, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP()),
('专业版', '更多功能', 2999, 50000, 'month', 1, 1, UNIX_TIMESTAMP(), UNIX_TIMESTAMP());
