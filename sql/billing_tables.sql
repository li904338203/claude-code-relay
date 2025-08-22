-- 计费系统数据库表创建脚本
-- 执行时间：请在数据库维护窗口执行

-- 1. 用户余额表
CREATE TABLE user_balances (
    id INT PRIMARY KEY AUTO_INCREMENT,
    user_id INT NOT NULL,
    balance DECIMAL(10,4) DEFAULT 0.0000 COMMENT '美元余额',
    frozen_balance DECIMAL(10,4) DEFAULT 0.0000 COMMENT '冻结余额',
    total_recharged DECIMAL(10,4) DEFAULT 0.0000 COMMENT '累计充值金额',
    total_consumed DECIMAL(10,4) DEFAULT 0.0000 COMMENT '累计消费金额',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_user_id (user_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) COMMENT='用户余额表';

-- 2. 充值卡表
CREATE TABLE recharge_cards (
    id INT PRIMARY KEY AUTO_INCREMENT,
    card_code VARCHAR(50) UNIQUE NOT NULL COMMENT '卡密',
    card_type ENUM('usage_count', 'time_limit', 'balance') NOT NULL COMMENT '卡类型：次数卡/时间卡/余额卡',
    
    -- 次数卡字段
    usage_count INT DEFAULT 0 COMMENT '可用次数（次数卡）',
    
    -- 时间卡字段
    time_type ENUM('daily', 'weekly', 'monthly') COMMENT '时间类型',
    duration_days INT DEFAULT 0 COMMENT '有效天数',
    daily_limit INT DEFAULT 0 COMMENT '每日使用限制',
    
    -- 通用字段
    value DECIMAL(10,4) NOT NULL COMMENT '面值（美元）',
    status ENUM('unused', 'used', 'expired') DEFAULT 'unused',
    user_id INT COMMENT '使用用户ID',
    used_at DATETIME COMMENT '使用时间',
    expired_at DATETIME COMMENT '过期时间',
    batch_id VARCHAR(50) COMMENT '批次ID',
    created_by INT COMMENT '创建者用户ID',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_card_code (card_code),
    INDEX idx_status (status),
    INDEX idx_batch_id (batch_id),
    INDEX idx_user_id (user_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL
) COMMENT='充值卡表';

-- 3. 用户卡套餐表
CREATE TABLE user_card_plans (
    id INT PRIMARY KEY AUTO_INCREMENT,
    user_id INT NOT NULL,
    card_id INT NOT NULL COMMENT '关联充值卡ID',
    plan_type ENUM('usage_count', 'time_limit') NOT NULL,
    
    -- 次数卡状态
    total_usage INT DEFAULT 0 COMMENT '总次数',
    used_usage INT DEFAULT 0 COMMENT '已用次数',
    remaining_usage INT DEFAULT 0 COMMENT '剩余次数',
    
    -- 时间卡状态
    time_type ENUM('daily', 'weekly', 'monthly'),
    daily_limit INT DEFAULT 0,
    start_date DATE COMMENT '开始日期',
    end_date DATE COMMENT '结束日期',
    today_used INT DEFAULT 0 COMMENT '今日已用',
    last_reset_date DATE COMMENT '上次重置日期',
    
    status ENUM('active', 'expired', 'exhausted') DEFAULT 'active',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_user_id (user_id),
    INDEX idx_status (status),
    INDEX idx_plan_type (plan_type),
    INDEX idx_end_date (end_date),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (card_id) REFERENCES recharge_cards(id) ON DELETE RESTRICT
) COMMENT='用户卡套餐表';

-- 4. 消费记录表
CREATE TABLE consumption_logs (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id INT NOT NULL,
    plan_id INT COMMENT '关联套餐ID',
    request_id VARCHAR(50) COMMENT '请求ID',
    api_key_id INT COMMENT 'API Key ID',
    account_id INT COMMENT '账号ID',
    
    -- 消费信息
    cost_usd DECIMAL(10,6) NOT NULL COMMENT '消费美元',
    usage_count INT DEFAULT 1 COMMENT '消费次数',
    deduction_type ENUM('balance', 'usage_count', 'time_limit') NOT NULL COMMENT '扣费类型',
    
    -- Token统计
    input_tokens INT DEFAULT 0,
    output_tokens INT DEFAULT 0,
    cache_read_tokens INT DEFAULT 0,
    cache_creation_tokens INT DEFAULT 0,
    total_tokens INT DEFAULT 0,
    
    model VARCHAR(100),
    platform_type VARCHAR(50),
    is_stream BOOLEAN DEFAULT FALSE,
    
    -- 余额变动记录
    balance_before DECIMAL(10,4) COMMENT '扣费前余额',
    balance_after DECIMAL(10,4) COMMENT '扣费后余额',
    
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    
    INDEX idx_user_id (user_id),
    INDEX idx_request_id (request_id),
    INDEX idx_api_key_id (api_key_id),
    INDEX idx_account_id (account_id),
    INDEX idx_created_at (created_at),
    INDEX idx_deduction_type (deduction_type),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (plan_id) REFERENCES user_card_plans(id) ON DELETE SET NULL,
    FOREIGN KEY (api_key_id) REFERENCES api_keys(id) ON DELETE SET NULL,
    FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE SET NULL
) COMMENT='消费记录表';

-- 5. 充值记录表
CREATE TABLE recharge_logs (
    id INT PRIMARY KEY AUTO_INCREMENT,
    user_id INT NOT NULL,
    card_id INT COMMENT '充值卡ID',
    amount DECIMAL(10,4) NOT NULL COMMENT '充值金额',
    recharge_type ENUM('card', 'manual', 'system') NOT NULL COMMENT '充值类型',
    description TEXT COMMENT '充值说明',
    operator_id INT COMMENT '操作员ID（管理员充值时）',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    
    INDEX idx_user_id (user_id),
    INDEX idx_recharge_type (recharge_type),
    INDEX idx_created_at (created_at),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (card_id) REFERENCES recharge_cards(id) ON DELETE SET NULL,
    FOREIGN KEY (operator_id) REFERENCES users(id) ON DELETE SET NULL
) COMMENT='充值记录表';

-- 初始化已有用户的余额记录
INSERT INTO user_balances (user_id, balance, created_at, updated_at)
SELECT id, 0.0000, NOW(), NOW() 
FROM users 
WHERE id NOT IN (SELECT user_id FROM user_balances);

-- 添加配置表（可选）
CREATE TABLE billing_config (
    id INT PRIMARY KEY AUTO_INCREMENT,
    config_key VARCHAR(100) UNIQUE NOT NULL,
    config_value TEXT NOT NULL,
    description VARCHAR(255) COMMENT '配置说明',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) COMMENT='计费配置表';

-- 插入默认配置
INSERT INTO billing_config (config_key, config_value, description) VALUES
('enable_usage_cards', 'true', '启用次数卡'),
('enable_time_cards', 'true', '启用时间卡'),
('min_balance', '0.0000', '最低余额限制'),
('max_daily_usage', '1000', '最大每日使用限制'),
('card_expire_days', '365', '卡默认有效期（天）'),
('enable_balance_alert', 'true', '启用余额不足提醒'),
('balance_alert_threshold', '1.0000', '余额预警阈值');