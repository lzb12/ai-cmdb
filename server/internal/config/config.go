package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	App   AppConfig   `mapstructure:"app"`
	MySQL MySQLConfig `mapstructure:"mysql"`
	Redis RedisConfig `mapstructure:"redis"`
	Auth  AuthConfig  `mapstructure:"auth"`
	Log   LogConfig   `mapstructure:"log"`
}

type AppConfig struct {
	Name string `mapstructure:"name"`
	Port string `mapstructure:"port"`
	Mode string `mapstructure:"mode"`
}

type MySQLConfig struct {
	Host                   string `mapstructure:"host"`
	Port                   int    `mapstructure:"port"`
	Username               string `mapstructure:"username"`
	Password               string `mapstructure:"password"`
	DB                     string `mapstructure:"db"`
	Charset                string `mapstructure:"charset"`
	ParseTime              bool   `mapstructure:"parse_time"`
	Loc                    string `mapstructure:"loc"`
	MaxIdleConns           int    `mapstructure:"max_idle_conns"`
	MaxOpenConns           int    `mapstructure:"max_open_conns"`
	ConnMaxLifetimeMinutes int    `mapstructure:"conn_max_lifetime_minutes"`
}

type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type AuthConfig struct {
	Secret          string `mapstructure:"secret"`
	SessionTTLHours int    `mapstructure:"session_ttl_hours"`
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

func Load() (*Config, error) {
	v := viper.New()
	v.SetConfigName("app")
	v.SetConfigType("toml")
	v.AddConfigPath(".")
	v.AddConfigPath("./config")
	v.AddConfigPath("./server")
	v.AddConfigPath("../")
	v.AddConfigPath("../../")

	v.SetEnvPrefix("CMDB")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	setDefaults(v)

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read app.toml failed: %w", err)
	}

	cfg := &Config{}
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("decode config failed: %w", err)
	}
	return cfg, nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("app.name", "cmdb-server")
	v.SetDefault("app.port", "8080")
	v.SetDefault("app.mode", "release")

	v.SetDefault("mysql.host", "127.0.0.1")
	v.SetDefault("mysql.port", 3306)
	v.SetDefault("mysql.username", "root")
	v.SetDefault("mysql.password", "root")
	v.SetDefault("mysql.db", "cmdb")
	v.SetDefault("mysql.charset", "utf8mb4")
	v.SetDefault("mysql.parse_time", true)
	v.SetDefault("mysql.loc", "Local")
	v.SetDefault("mysql.max_idle_conns", 10)
	v.SetDefault("mysql.max_open_conns", 50)
	v.SetDefault("mysql.conn_max_lifetime_minutes", 60)

	v.SetDefault("redis.addr", "127.0.0.1:6379")
	v.SetDefault("redis.password", "")
	v.SetDefault("redis.db", 0)

	v.SetDefault("auth.secret", "change-me-to-a-strong-secret-key")
	v.SetDefault("auth.session_ttl_hours", 12)

	v.SetDefault("log.level", "info")
	v.SetDefault("log.filename", "logs/app.log")
	v.SetDefault("log.max_size", 20)
	v.SetDefault("log.max_backups", 10)
	v.SetDefault("log.max_age", 30)
	v.SetDefault("log.compress", true)
	v.SetDefault("log.console", true)
}
