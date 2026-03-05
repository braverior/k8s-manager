package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Server     ServerConfig     `mapstructure:"server"`
	Database   DatabaseConfig   `mapstructure:"database"`
	Log        LogConfig        `mapstructure:"log"`
	Feishu     FeishuConfig     `mapstructure:"feishu"`
	JWT        JWTConfig        `mapstructure:"jwt"`
	Encryption EncryptionConfig `mapstructure:"encryption"`
}

type ServerConfig struct {
	Port int    `mapstructure:"port"`
	Mode string `mapstructure:"mode"`
}

type DatabaseConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	Username     string `mapstructure:"username"`
	Password     string `mapstructure:"password"`
	Database     string `mapstructure:"database"`
	Charset      string `mapstructure:"charset"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
}

type LogConfig struct {
	Level    string `mapstructure:"level"`
	Format   string `mapstructure:"format"`
	Output   string `mapstructure:"output"`
	FilePath string `mapstructure:"file_path"`
}

// FeishuConfig 飞书配置
type FeishuConfig struct {
	AppID        string `mapstructure:"app_id"`        // 飞书应用 App ID
	AppSecret    string `mapstructure:"app_secret"`    // 飞书应用 App Secret
	RedirectURI  string `mapstructure:"redirect_uri"`  // 授权回调地址（前端地址）
	AuthorizeURL string `mapstructure:"authorize_url"` // 飞书授权页面 URL
}

// JWTConfig JWT 配置
type JWTConfig struct {
	Secret     string `mapstructure:"secret"`      // JWT 签名密钥
	ExpireTime int    `mapstructure:"expire_time"` // 过期时间（秒）
	Issuer     string `mapstructure:"issuer"`      // 签发者
}

// EncryptionConfig 加密配置
type EncryptionConfig struct {
	Key string `mapstructure:"key"` // AES 加密密钥
}

func (d *DatabaseConfig) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
		d.Username, d.Password, d.Host, d.Port, d.Database, d.Charset)
}

var globalConfig *Config

func Load(configPath string) (*Config, error) {
	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")

	// Support environment variable override
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// 绑定环境变量（用于敏感配置）
	viper.BindEnv("feishu.app_id", "FEISHU_APP_ID")
	viper.BindEnv("feishu.app_secret", "FEISHU_APP_SECRET")
	viper.BindEnv("feishu.redirect_uri", "FEISHU_REDIRECT_URI")
	viper.BindEnv("jwt.secret", "JWT_SECRET")
	viper.BindEnv("database.password", "DATABASE_PASSWORD")
	viper.BindEnv("encryption.key", "ENCRYPTION_KEY")

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// 设置默认值
	if cfg.Feishu.AuthorizeURL == "" {
		cfg.Feishu.AuthorizeURL = "https://passport.feishu.cn/suite/passport/oauth/authorize"
	}
	if cfg.JWT.ExpireTime == 0 {
		cfg.JWT.ExpireTime = 86400 // 默认 24 小时
	}
	if cfg.JWT.Issuer == "" {
		cfg.JWT.Issuer = "k8s-manager"
	}
	if cfg.Encryption.Key == "" {
		cfg.Encryption.Key = "k8s-manager-default-encryption-key"
	}

	globalConfig = &cfg
	return &cfg, nil
}

func Get() *Config {
	return globalConfig
}
