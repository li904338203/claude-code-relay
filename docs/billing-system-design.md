# Claude Code Relay 计费系统设计文档

## 项目现状分析

### 当前计费系统
1. **费用计算**：已有完整的Token计费系统（`common/cost_calculator.go`）
   - 按模型定价（Input/Output tokens，缓存读写）
   - 支持多种Claude模型定价
   - 以美元（USD）为单位计算费用

2. **账户管理**：Account模型已有今日使用统计
   - `TodayUsageCount`：今日使用次数
   - `TodayTotalCost`：今日总费用（USD）
   - `TodayInputTokens`、`TodayOutputTokens`等

3. **用户系统**：基础用户管理，但缺少余额/充值功能

## 功能设计架构

### 1. 数据库表设计

#### 用户余额表 (user_balances)
```sql
CREATE TABLE user_balances (
    id INT PRIMARY KEY AUTO_INCREMENT,
    user_id INT NOT NULL,
    balance DECIMAL(10,4) DEFAULT 0.0000 COMMENT '美元余额',
    frozen_balance DECIMAL(10,4) DEFAULT 0.0000 COMMENT '冻结余额',
    created_at DATETIME,
    updated_at DATETIME
);
```

#### 充值卡表 (recharge_cards)
```sql
CREATE TABLE recharge_cards (
    id INT PRIMARY KEY AUTO_INCREMENT,
    card_code VARCHAR(50) UNIQUE NOT NULL COMMENT '卡密',
    card_type ENUM('usage_count', 'time_limit') NOT NULL COMMENT '卡类型',
    
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
    created_at DATETIME,
    updated_at DATETIME
);
```

#### 用户卡套餐表 (user_card_plans)
```sql
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
    created_at DATETIME,
    updated_at DATETIME
);
```

#### 消费记录表 (consumption_logs)
```sql
CREATE TABLE consumption_logs (
    id INT PRIMARY KEY AUTO_INCREMENT,
    user_id INT NOT NULL,
    plan_id INT COMMENT '关联套餐ID',
    request_id VARCHAR(50) COMMENT '请求ID',
    
    -- 消费信息
    cost_usd DECIMAL(10,6) NOT NULL COMMENT '消费美元',
    usage_count INT DEFAULT 1 COMMENT '消费次数',
    deduction_type ENUM('balance', 'usage_count', 'time_limit') NOT NULL,
    
    -- Token统计
    input_tokens INT DEFAULT 0,
    output_tokens INT DEFAULT 0,
    cache_read_tokens INT DEFAULT 0,
    cache_creation_tokens INT DEFAULT 0,
    
    model VARCHAR(100),
    created_at DATETIME
);
```

### 2. 核心业务逻辑设计

#### 计费优先级策略
```
1. 时间卡套餐（在有效期内且未达每日限制）
2. 次数卡套餐（有剩余次数）
3. 余额扣费（账户余额充足）
4. 拒绝服务（无可用资源）
```

#### 次数卡设计
- **核心概念**：预付费买入固定调用次数
- **计费方式**：每次API调用消费1次，不考虑实际token费用
- **适用场景**：轻度用户，预算控制
- **过期机制**：可设置有效期，过期后作废

#### 时间卡设计
- **核心概念**：在指定时间段内每日限制使用次数
- **时间类型**：
  - 日卡：24小时内限制N次调用
  - 周卡：7天内每天限制N次调用  
  - 月卡：30天内每天限制N次调用
- **重置机制**：每日0点重置当天使用次数
- **费用模式**：套餐内免费使用，超出部分按余额扣费

### 3. API接口设计

#### 管理员接口
```
POST /admin/cards/generate          # 生成充值卡
GET  /admin/cards                   # 查询充值卡
PUT  /admin/cards/:id/status        # 修改卡状态
GET  /admin/consumption/stats       # 消费统计
```

#### 用户接口
```
POST /user/cards/redeem            # 充值卡兑换
GET  /user/balance                 # 查询余额和套餐
GET  /user/plans                   # 查询我的套餐
GET  /user/consumption/history     # 消费历史
POST /user/balance/recharge        # 余额充值（如需要）
```

#### 内部计费接口
```
POST /internal/billing/deduct      # 扣费处理
GET  /internal/billing/check       # 检查用户可用额度
POST /internal/billing/record      # 记录消费日志
```

### 4. 计费流程设计

#### 请求前检查流程
```
1. 检查用户时间卡套餐
   - 是否在有效期内
   - 今日是否达到限制
   - 如果可用，允许请求

2. 检查用户次数卡套餐  
   - 是否有剩余次数
   - 如果可用，允许请求

3. 检查用户余额
   - 预估请求费用
   - 余额是否充足
   - 如果充足，冻结预估费用

4. 全部不满足，拒绝请求
```

#### 请求后扣费流程
```
1. 计算实际费用
2. 按优先级扣费：
   - 时间卡：记录使用次数，今日已用+1
   - 次数卡：剩余次数-1，记录消费
   - 余额扣费：解冻预估费用，扣除实际费用
3. 记录消费日志
4. 更新用户统计
```

### 5. 关键技术实现点

#### 并发安全
- 使用Redis锁确保扣费操作原子性
- 数据库事务保证数据一致性

#### 费用计算集成
- 复用现有`cost_calculator.go`
- 扩展支持套餐费用计算

#### 定时任务
- 每日重置时间卡使用次数
- 清理过期卡和套餐
- 统计报表生成

#### 监控告警
- 用户余额不足提醒
- 套餐即将到期通知
- 异常消费监控

### 6. 配置管理

#### 系统配置
```go
type BillingConfig struct {
    EnableUsageCards bool     // 启用次数卡
    EnableTimeCards  bool     // 启用时间卡
    MinBalance      float64   // 最低余额限制
    MaxDailyUsage   int       // 最大每日使用限制
    CardExpireDays  int       // 卡默认有效期
}
```

## 三种计费方式的关联逻辑

### 1. 基本关联模式

**层级关系**：
```
时间卡套餐 > 次数卡套餐 > 余额扣费
```

这是一个**优先级递减**的关联关系，用户可以同时拥有多种计费方式，系统按优先级自动选择最优惠的方案。

### 2. 具体关联场景

#### 场景A：纯时间卡用户
```
用户状态：
- 时间卡：月卡，每日100次限制，剩余15天
- 次数卡：无
- 余额：$0

扣费逻辑：
- 每日前100次：免费使用（时间卡套餐内）
- 超出100次：拒绝服务（余额不足）
```

#### 场景B：时间卡 + 余额用户
```
用户状态：
- 时间卡：月卡，每日50次限制，剩余20天  
- 次数卡：无
- 余额：$10.00

扣费逻辑：
- 每日前50次：免费使用（时间卡套餐内）
- 第51次起：按实际token费用扣除余额
- 余额用完：拒绝服务
```

#### 场景C：次数卡 + 余额用户
```
用户状态：
- 时间卡：无
- 次数卡：1000次，剩余800次
- 余额：$5.00

扣费逻辑：
- 前800次：按次数卡扣费（每次消费1次数，不扣美元）
- 第801次起：按实际token费用扣除余额
- 余额用完：拒绝服务
```

#### 场景D：全套餐用户
```
用户状态：
- 时间卡：周卡，每日20次限制，剩余3天
- 次数卡：500次，剩余200次  
- 余额：$15.00

扣费逻辑：
1. 每日前20次：优先使用时间卡（免费）
2. 第21次起：使用次数卡（消费次数，不扣美元）
3. 次数卡用完后：扣除余额
4. 时间卡到期后：优先使用次数卡，然后余额
```

### 3. 购买和充值关联

#### 充值卡兑换逻辑
```
次数卡兑换：
- 用户兑换1000次卡（面值$30）
- 系统创建user_card_plans记录
- 不直接增加用户美元余额
- 按次消费，每次调用-1次数

时间卡兑换：
- 用户兑换月卡（面值$50，每日100次）
- 系统创建user_card_plans记录  
- 设置30天有效期和每日限制
- 套餐内使用不扣费

余额充值：
- 用户兑换余额卡（面值$20）
- 直接增加user_balances.balance
- 按实际token消费扣费
```

### 4. 成本核算关联

#### 对用户的计费策略
```
时间卡：固定套餐费，套餐内免费使用
次数卡：固定次数费，按次计费不论实际消费
余额：按量付费，根据实际token消费扣费
```

#### 对平台的成本计算
```
每次API调用，平台实际成本：
- 根据token数量和模型计算真实美元成本
- 记录在consumption_logs表中

收益分析：
- 时间卡：预收费模式，用户用量越少平台收益越高
- 次数卡：预收费模式，重度用户可能亏损，轻度用户盈利
- 余额：即付即收，基本无风险
```

### 5. 数据流转关联

#### 用户维度数据关联
```sql
-- 用户完整计费状态查询
SELECT 
    u.id,
    u.username,
    ub.balance,                    -- 余额
    COUNT(ucp1.id) as time_cards,  -- 时间卡数量
    COUNT(ucp2.id) as usage_cards, -- 次数卡数量
    SUM(cl.cost_usd) as total_spent -- 总消费
FROM users u
LEFT JOIN user_balances ub ON u.id = ub.user_id
LEFT JOIN user_card_plans ucp1 ON u.id = ucp1.user_id AND ucp1.plan_type = 'time_limit'
LEFT JOIN user_card_plans ucp2 ON u.id = ucp2.user_id AND ucp2.plan_type = 'usage_count'  
LEFT JOIN consumption_logs cl ON u.id = cl.user_id
GROUP BY u.id;
```

#### 扣费决策流程图
```
API请求到达
    ↓
检查用户时间卡套餐
    ↓
[有效时间卡] → 今日限制内？ → 是 → 免费使用，记录消费
    ↓ 否                           ↓
检查用户次数卡套餐               检查次数卡
    ↓                               ↓
[有剩余次数] → 是 → 消费1次数，记录消费
    ↓ 否                           ↓
检查用户余额                    检查余额
    ↓                               ↓
[余额充足] → 是 → 扣除实际费用，记录消费
    ↓ 否                           ↓
拒绝服务 ← ← ← ← ← ← ← ← ← ← ← 拒绝服务
```

### 6. 业务规则关联

#### 套餐叠加规则
- **时间卡**：可以同时拥有多张不同时效的时间卡（日卡+周卡+月卡）
- **次数卡**：可以叠加，总次数累计
- **余额**：统一账户，可以多次充值累计

#### 优先级规则
```
1. 时间卡优先级：日卡 > 周卡 > 月卡（先用短期的）
2. 次数卡优先级：先到期的优先使用
3. 余额：统一池子，无优先级
```

#### 退费和转换规则
```
时间卡 → 不支持退费，不支持转余额
次数卡 → 不支持退费，可考虑转余额（按比例）
余额   → 可以提现（需要手续费和审核）
```

## 总结

这种设计让三种计费方式既相互独立又有机结合，用户可以根据使用习惯选择最适合的付费方式，平台也能通过多样化的套餐获得更稳定的现金流。同时保持了与现有系统的兼容性，提供了灵活的计费方案，可以满足不同用户的需求场景。