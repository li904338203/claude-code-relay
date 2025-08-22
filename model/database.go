package model

import (
	"claude-code-relay/common"
	"fmt"
	"os"
	"strconv"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func InitDB() error {
	// MySQL 数据库连接配置
	host := os.Getenv("MYSQL_HOST")
	if host == "" {
		host = "localhost"
	}

	port := os.Getenv("MYSQL_PORT")
	if port == "" {
		port = "3306"
	}

	user := os.Getenv("MYSQL_USER")
	if user == "" {
		user = "root"
	}

	password := os.Getenv("MYSQL_PASSWORD")
	if password == "" {
		password = ""
	}

	database := os.Getenv("MYSQL_DATABASE")
	if database == "" {
		database = "claude_code_relay"
	}

	// 先连接到MySQL服务器（不指定数据库）
	adminDsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/?charset=utf8mb4&parseTime=True&loc=Local",
		user, password, host, port)

	adminDB, err := gorm.Open(mysql.Open(adminDsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return fmt.Errorf("failed to connect to MySQL server: %v", err)
	}

	// 创建数据库（如果不存在）
	createDBSQL := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci", database)
	if err := adminDB.Exec(createDBSQL).Error; err != nil {
		return fmt.Errorf("failed to create database: %v", err)
	}

	// 关闭管理连接
	adminSqlDB, _ := adminDB.DB()
	adminSqlDB.Close()

	// 构建应用数据库 DSN
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		user, password, host, port, database)

	// 连接到应用数据库
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return fmt.Errorf("failed to connect to application database: %v", err)
	}

	// 配置数据库连接池
	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}

	// 设置最大打开连接数（默认100）
	maxOpenConns := getIntEnv("MYSQL_MAX_OPEN_CONNS", 100)
	sqlDB.SetMaxOpenConns(maxOpenConns)

	// 设置最大空闲连接数（默认10）
	maxIdleConns := getIntEnv("MYSQL_MAX_IDLE_CONNS", 10)
	sqlDB.SetMaxIdleConns(maxIdleConns)

	// 设置连接最大生存时间（默认1小时）
	maxLifetimeMinutes := getIntEnv("MYSQL_MAX_LIFETIME_MINUTES", 60)
	sqlDB.SetConnMaxLifetime(time.Duration(maxLifetimeMinutes) * time.Minute)

	// 设置连接最大空闲时间（默认30分钟）
	maxIdleTimeMinutes := getIntEnv("MYSQL_MAX_IDLE_TIME_MINUTES", 30)
	sqlDB.SetConnMaxIdleTime(time.Duration(maxIdleTimeMinutes) * time.Minute)

	// 自动迁移数据库表
	err = DB.AutoMigrate(
		&User{},
		&Task{},
		&ApiLog{},
		&Account{},
		&Group{},
		&ApiKey{},
		&Log{},
		// 计费系统相关表
		&UserBalance{},
		&RechargeCard{},
		&UserCardPlan{},
		&ConsumptionLog{},
		&RechargeLog{},
		&BillingConfig{},
		// 模型配置相关表
		&ModelConfig{},
		&ModelPricing{},
	)
	if err != nil {
		return err
	}

	// 初始化计费系统默认配置
	err = initBillingData()
	if err != nil {
		common.SysLog("Warning: Failed to initialize billing data: " + err.Error())
	}

	
	// 初始化模型配置数据
	err = initModelData()
	if err != nil {
		common.SysLog("Warning: Failed to initialize model data: " + err.Error())
	}

	// 初始化模型定价服务
	InitializeModelPricingService()

	common.SysLog("Database initialized successfully")
	return nil
}

func CloseDB() error {
	if DB != nil {
		sqlDB, err := DB.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}

// getIntEnv 获取环境变量的整型值，如果不存在或无效则返回默认值
func getIntEnv(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	intValue, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}

	return intValue
}

// initBillingData 初始化计费系统数据
func initBillingData() error {
	// 初始化计费配置
	defaultConfigs := []BillingConfig{
		{ConfigKey: "enable_usage_cards", ConfigValue: "true", Description: stringPtr("启用次数卡")},
		{ConfigKey: "enable_time_cards", ConfigValue: "true", Description: stringPtr("启用时间卡")},
		{ConfigKey: "min_balance", ConfigValue: "0.0000", Description: stringPtr("最低余额限制")},
		{ConfigKey: "max_daily_usage", ConfigValue: "1000", Description: stringPtr("最大每日使用限制")},
		{ConfigKey: "card_expire_days", ConfigValue: "365", Description: stringPtr("卡默认有效期（天）")},
		{ConfigKey: "enable_balance_alert", ConfigValue: "true", Description: stringPtr("启用余额不足提醒")},
		{ConfigKey: "balance_alert_threshold", ConfigValue: "1.0000", Description: stringPtr("余额预警阈值")},
	}

	for _, config := range defaultConfigs {
		var existingConfig BillingConfig
		result := DB.Where("config_key = ?", config.ConfigKey).First(&existingConfig)
		if result.Error != nil {
			// 配置不存在，创建新配置
			if err := DB.Create(&config).Error; err != nil {
				return fmt.Errorf("failed to create billing config %s: %v", config.ConfigKey, err)
			}
		}
	}

	// 为现有用户创建余额记录
	var users []User
	if err := DB.Find(&users).Error; err != nil {
		return fmt.Errorf("failed to fetch users: %v", err)
	}

	for _, user := range users {
		var existingBalance UserBalance
		result := DB.Where("user_id = ?", user.ID).First(&existingBalance)
		if result.Error != nil {
			// 用户余额记录不存在，创建默认余额记录
			balance := UserBalance{
				UserID:         user.ID,
				Balance:        0.0000,
				FrozenBalance:  0.0000,
				TotalRecharged: 0.0000,
				TotalConsumed:  0.0000,
			}
			if err := DB.Create(&balance).Error; err != nil {
				return fmt.Errorf("failed to create balance for user %d: %v", user.ID, err)
			}
		}
	}

	return nil
}

// stringPtr 返回字符串指针
func stringPtr(s string) *string {
	return &s
}

// initModelData 初始化模型配置数据
func initModelData() error {
	// 检查是否已经有模型配置数据
	var count int64
	if err := DB.Model(&ModelConfig{}).Count(&count).Error; err != nil {
		return fmt.Errorf("failed to count model configs: %v", err)
	}

	// 如果已经有数据，跳过初始化
	if count > 0 {
		common.SysLog("Model configurations already exist, skipping initialization")
		return nil
	}

	common.SysLog("Initializing model configurations from hardcoded data...")

	// 从硬编码数据初始化模型配置
	modelConfigs := []ModelConfig{
		{
			Name:        "claude-3-5-sonnet-20241022",
			DisplayName: "Claude 3.5 Sonnet",
			Provider:    "claude",
			Category:    "sonnet",
			Version:     "20241022",
			Status:      1,
			SortOrder:   1,
			Description: "Claude 3.5 Sonnet - 高性能的AI助手模型",
		},
		{
			Name:        "claude-sonnet-4-20250514",
			DisplayName: "Claude Sonnet 4",
			Provider:    "claude",
			Category:    "sonnet",
			Version:     "20250514",
			Status:      1,
			SortOrder:   2,
			Description: "Claude Sonnet 4 - 下一代Sonnet模型",
		},
		{
			Name:        "claude-opus-4-20250514",
			DisplayName: "Claude Opus 4",
			Provider:    "claude",
			Category:    "opus",
			Version:     "20250514",
			Status:      1,
			SortOrder:   3,
			Description: "Claude Opus 4 - 最强性能的AI助手模型",
		},
		{
			Name:        "claude-opus-4-1-20250805",
			DisplayName: "Claude Opus 4.1",
			Provider:    "claude",
			Category:    "opus",
			Version:     "20250805",
			Status:      1,
			SortOrder:   4,
			Description: "Claude Opus 4.1 - 增强版Opus模型",
		},
		{
			Name:        "claude-3-5-haiku-20241022",
			DisplayName: "Claude 3.5 Haiku",
			Provider:    "claude",
			Category:    "haiku",
			Version:     "20241022",
			Status:      1,
			SortOrder:   5,
			Description: "Claude 3.5 Haiku - 快速响应的轻量级模型",
		},
		{
			Name:        "claude-3-opus-20240229",
			DisplayName: "Claude 3 Opus",
			Provider:    "claude",
			Category:    "opus",
			Version:     "20240229",
			Status:      1,
			SortOrder:   6,
			Description: "Claude 3 Opus - 强大的第三代Opus模型",
		},
		{
			Name:        "claude-3-sonnet-20240229",
			DisplayName: "Claude 3 Sonnet",
			Provider:    "claude",
			Category:    "sonnet",
			Version:     "20240229",
			Status:      1,
			SortOrder:   7,
			Description: "Claude 3 Sonnet - 均衡性能的第三代Sonnet模型",
		},
		{
			Name:        "claude-3-haiku-20240307",
			DisplayName: "Claude 3 Haiku",
			Provider:    "claude",
			Category:    "haiku",
			Version:     "20240307",
			Status:      1,
			SortOrder:   8,
			Description: "Claude 3 Haiku - 第三代快速响应模型",
		},
		{
			Name:        "unknown",
			DisplayName: "未知模型",
			Provider:    "claude",
			Category:    "default",
			Version:     "default",
			Status:      1,
			SortOrder:   999,
			Description: "用于未知模型的默认定价配置",
		},
	}

	// 创建模型配置
	for _, model := range modelConfigs {
		if err := DB.Create(&model).Error; err != nil {
			return fmt.Errorf("failed to create model config %s: %v", model.Name, err)
		}
	}

	// 获取创建的模型配置
	var createdModels []ModelConfig
	if err := DB.Find(&createdModels).Error; err != nil {
		return fmt.Errorf("failed to fetch created models: %v", err)
	}

	// 创建模型定价配置
	modelPricingData := map[string]struct {
		Input      float64
		Output     float64
		CacheWrite float64
		CacheRead  float64
	}{
		"claude-3-5-sonnet-20241022": {
			Input:      3.00,
			Output:     15.00,
			CacheWrite: 3.75,
			CacheRead:  0.30,
		},
		"claude-sonnet-4-20250514": {
			Input:      3.00,
			Output:     15.00,
			CacheWrite: 3.75,
			CacheRead:  0.30,
		},
		"claude-opus-4-20250514": {
			Input:      15.00,
			Output:     75.00,
			CacheWrite: 18.75,
			CacheRead:  1.50,
		},
		"claude-opus-4-1-20250805": {
			Input:      15.00,
			Output:     75.00,
			CacheWrite: 18.75,
			CacheRead:  1.50,
		},
		"claude-3-5-haiku-20241022": {
			Input:      0.25,
			Output:     1.25,
			CacheWrite: 0.30,
			CacheRead:  0.03,
		},
		"claude-3-opus-20240229": {
			Input:      15.00,
			Output:     75.00,
			CacheWrite: 18.75,
			CacheRead:  1.50,
		},
		"claude-3-sonnet-20240229": {
			Input:      3.00,
			Output:     15.00,
			CacheWrite: 3.75,
			CacheRead:  0.30,
		},
		"claude-3-haiku-20240307": {
			Input:      0.25,
			Output:     1.25,
			CacheWrite: 0.30,
			CacheRead:  0.03,
		},
		"unknown": {
			Input:      3.00,
			Output:     15.00,
			CacheWrite: 3.75,
			CacheRead:  0.30,
		},
	}

	// 为每个模型创建定价配置
	effectiveTime := Time(time.Now())
	for _, model := range createdModels {
		if pricingData, exists := modelPricingData[model.Name]; exists {
			pricing := ModelPricing{
				ModelID:         model.ID,
				InputPrice:      pricingData.Input,
				OutputPrice:     pricingData.Output,
				CacheWritePrice: pricingData.CacheWrite,
				CacheReadPrice:  pricingData.CacheRead,
				EffectiveTime:   effectiveTime,
				Status:          1,
			}
			if err := DB.Create(&pricing).Error; err != nil {
				return fmt.Errorf("failed to create pricing for model %s: %v", model.Name, err)
			}
		}
	}

	common.SysLog("Model configurations initialized successfully")
	return nil
}
