package database

import (
	"fmt"
	"time"

	"goboot/config"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func InitMySQL() error {
	cfg := config.AppConfig.MySQL
	// DSN 添加连接参数:
	// - timeout: 连接超时时间
	// - readTimeout: 读取超时
	// - writeTimeout: 写入超时
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local&timeout=10s&readTimeout=30s&writeTimeout=30s",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Database,
		cfg.Charset,
	)

	var logMode logger.LogLevel
	if config.AppConfig.Server.Mode == "debug" {
		logMode = logger.Info
	} else {
		logMode = logger.Silent
	}

	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logMode),
	})
	if err != nil {
		return err
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}

	// 设置连接池参数
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)

	// 设置连接最大生命周期(建议小于数据库的wait_timeout，MySQL默认8小时)
	// 超过此时间的连接会被关闭并重新创建
	sqlDB.SetConnMaxLifetime(time.Hour * 1)

	// 设置空闲连接最大生命周期
	// 空闲超过此时间的连接会被关闭
	sqlDB.SetConnMaxIdleTime(time.Minute * 30)

	return nil
}
