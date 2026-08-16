package config

import (
	"os"
)

type Config struct {
	Addr       string // 监听地址，如 :8080
	DBPath     string // SQLite 文件路径
	JWTSecret  string // 令牌签名密钥
	SeedOnBoot bool   // 首次启动时是否写入演示数据
}

func Load() Config {
	return Config{
		Addr:       getEnv("APP_ADDR", ":8080"),
		DBPath:     getEnv("APP_DB_PATH", "./data/approval.db"),
		JWTSecret:  getEnv("APP_JWT_SECRET", "dev-secret-change-me-in-production"),
		SeedOnBoot: getEnv("APP_SEED_ON_BOOT", "true") == "true",
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
