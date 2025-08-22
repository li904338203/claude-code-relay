-- 添加 disabled 状态到充值卡和用户套餐表的枚举字段
-- 需要手动执行这些SQL语句来更新数据库结构

-- 1. 更新充值卡表的 status 字段，添加 disabled 选项
ALTER TABLE recharge_cards 
MODIFY COLUMN status ENUM('unused', 'used', 'expired', 'disabled') DEFAULT 'unused';

-- 2. 更新用户套餐表的 status 字段，添加 disabled 选项  
ALTER TABLE user_card_plans 
MODIFY COLUMN status ENUM('active', 'expired', 'exhausted', 'disabled') DEFAULT 'active';

-- 3. 创建索引以提高查询效率（可选）
CREATE INDEX IF NOT EXISTS idx_recharge_cards_status ON recharge_cards(status);
CREATE INDEX IF NOT EXISTS idx_user_card_plans_status ON user_card_plans(status);

-- 4. 验证更新结果
SELECT COLUMN_NAME, COLUMN_TYPE 
FROM INFORMATION_SCHEMA.COLUMNS 
WHERE TABLE_NAME = 'recharge_cards' AND COLUMN_NAME = 'status';

SELECT COLUMN_NAME, COLUMN_TYPE 
FROM INFORMATION_SCHEMA.COLUMNS 
WHERE TABLE_NAME = 'user_card_plans' AND COLUMN_NAME = 'status';