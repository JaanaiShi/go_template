package redis

import (
	"fmt"
	"github.com/FZambia/sentinel"
	"github.com/gomodule/redigo/redis"
	"go.uber.org/zap"
	"strings"
	"time"
)

var (
	RedisConnPool *redis.Pool
	ErrNil        = "redigo: nil returned"
)

func InitRedis(c C) {
	redisAddr := c.RedisBase.RedisAddr
	redisType := c.RedisBase.RedisType
	MaxIdle := c.RedisBase.MaxIdle
	masterName := c.RedisBase.MasterName
	password := c.RedisBase.RedisPassword
	db := c.RedisBase.Db
	if redisType == "sentinel" {
		redisAddrs := strings.Split(redisAddr, ",")
		sntnl := &sentinel.Sentinel{
			Addrs:      redisAddrs,
			MasterName: masterName,
			Dial: func(addr string) (redis.Conn, error) {
				timeout := 500 * time.Millisecond
				c, err := redis.Dial("tcp", addr, redis.DialWriteTimeout(timeout))
				if err != nil {
					return nil, err
				}
				return c, nil
			},
		}

		RedisConnPool = &redis.Pool{
			MaxIdle:     MaxIdle,
			IdleTimeout: 240 * time.Second,
			Dial: func() (redis.Conn, error) {
				masterAddr, err := sntnl.MasterAddr()
				if err != nil {
					return nil, err
				}
				setdb := redis.DialDatabase(db)
				pd := redis.DialPassword(password)
				c, err := redis.Dial("tcp", masterAddr, setdb, pd)
				if err != nil {
					return nil, err
				}
				return c, nil
			},
			TestOnBorrow: CheckRedisRole,
		}
	} else {
		RedisConnPool = &redis.Pool{
			MaxIdle:     MaxIdle,
			IdleTimeout: 240 * time.Second,
			// Dial or DialContext must be set. When both are set, DialContext takes precedence over Dial.
			Dial: func() (redis.Conn, error) {
				setdb := redis.DialDatabase(db)
				pd := redis.DialPassword(password)
				c, err := redis.Dial("tcp", redisAddr, setdb, pd)
				if err != nil {
					c.Close()
					GVA_LOG.Info("初始化redis失败", zap.String("status", "失败"))
					panic(err)
				}
				return c, nil
			},
		}
	}
}

func CheckRedisRole(c redis.Conn, t time.Time) error {
	if !sentinel.TestRole(c, "master") {
		return fmt.Errorf("Role check failed")
	} else {
		return nil
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
