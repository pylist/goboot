package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	Server    ServerConfig    `mapstructure:"server"`
	MySQL     MySQLConfig     `mapstructure:"mysql"`
	Redis     RedisConfig     `mapstructure:"redis"`
	JWT       JWTConfig       `mapstructure:"jwt"`
	Log       LogConfig       `mapstructure:"log"`
	RateLimit RateLimitConfig `mapstructure:"rate_limit"`
	Email     EmailConfig     `mapstructure:"email"`
}

type ServerConfig struct {
	Host           string   `mapstructure:"host"`
	Port           int      `mapstructure:"port"`
	Mode           string   `mapstructure:"mode"`
	TrustedProxies []string `mapstructure:"trusted_proxies"` // 可信代理IP列表，空则不信任任何代理
}

type MySQLConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	User         string `mapstructure:"user"`
	Password     string `mapstructure:"password"`
	Database     string `mapstructure:"database"`
	Charset      string `mapstructure:"charset"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
	PoolSize int    `mapstructure:"pool_size"`
}

type JWTConfig struct {
	Secret        string `mapstructure:"secret"`
	AccessExpire  int    `mapstructure:"access_expire"`  // Access Token过期时间(小时)
	RefreshExpire int    `mapstructure:"refresh_expire"` // Refresh Token过期时间(小时)
	RefreshSecret string `mapstructure:"refresh_secret"` // Refresh Token密钥
}

type LogConfig struct {
	Level      string `mapstructure:"level"`
	Filename   string `mapstructure:"filename"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxBackups int    `mapstructure:"max_backups"`
	MaxAge     int    `mapstructure:"max_age"`
	Compress   bool   `mapstructure:"compress"`
	Console    bool   `mapstructure:"console"`
}

type RateLimitConfig struct {
	Enabled  bool `mapstructure:"enabled"`  // 是否启用限流
	Requests int  `mapstructure:"requests"` // 时间窗口内允许的请求数
	Window   int  `mapstructure:"window"`   // 时间窗口（秒）
}

type EmailConfig struct {
	Enabled    bool   `mapstructure:"enabled"`     // 是否启用邮件服务
	Host       string `mapstructure:"host"`        // SMTP 服务器地址
	Port       int    `mapstructure:"port"`        // SMTP 端口
	Username   string `mapstructure:"username"`    // 邮箱账号
	Password   string `mapstructure:"password"`    // 邮箱密码或授权码
	FromName   string `mapstructure:"from_name"`   // 发件人名称
	FromAddr   string `mapstructure:"from_addr"`   // 发件人地址
	SSL        bool   `mapstructure:"ssl"`         // 是否启用 SSL
	ResetURL   string `mapstructure:"reset_url"`   // 密码重置页面地址
	ResetExpire int   `mapstructure:"reset_expire"` // 重置链接过期时间（分钟）
}

var AppConfig *Config

func InitConfig() error {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		return err
	}

	AppConfig = &Config{}
	if err := viper.Unmarshal(AppConfig); err != nil {
		return err
	}

	return nil
}
