package common

import (
	"github.com/gomodule/redigo/redis"
	"gorm.io/gorm"
)

var (
	Conf          *Config
	DB            *gorm.DB
	RedisConnPool *redis.Pool
)

type Config struct {
	System struct {
		LogLevel string `json:"log_level"`
		LogPath  string `json:"log_path"`
	} `json:"system"`
	DB struct {
		Host      string `json:"host"`
		Port      string `json:"port"`
		Name      string `json:"name"`
		Username  string `json:"username"`
		Password  string `json:"password"`
		Config    string `json:"config"`
		MaxIdle   int    `json:"max_idle"`
		MaxOpen   int    `json:"max_open"`
		DbLogMode bool   `json:"db_log_mode"`
		LogZap    string `json:"log_zap"` // 留空不写到日志文件，gorm日志级别："silent", "Silent"  ｜  "error", "Error"  ｜ "warn", "Warn" ｜ "info", "Info" ｜ "zap", "Zap"
	}
	Redis struct {
		Host       string `json:"host"`
		Port       int    `json:"port"`
		MaxIdle    int    `json:"max_idle"`
		MasterName string `json:"master_name"`
		RedisType  string `json:"redis_type"`
		Password   string `json:"password"`
		Db         int    `json:"db"`
	}
}

func GetRedisPool() *redis.Pool {
	return RedisConnPool
}

func GetRedis() (redis.Conn, error) {
	conn := RedisConnPool.Get()
	err := conn.Err()
	return conn, err
}

func GetDB() *gorm.DB {
	if DB == nil {
		panic("please init db first")
	}
	return DB
}
