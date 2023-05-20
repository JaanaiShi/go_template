package redis

import (
	"fmt"
	"github.com/FZambia/sentinel"
	"github.com/JaanaiShi/flint/common"
	"github.com/gomodule/redigo/redis"
	"strconv"
	"strings"
	"time"
)

var (
	ErrNil = "redigo: nil returned"
)

func Init() {
	redisAddr := common.Conf.Redis.Host + ":" + strconv.Itoa(common.Conf.Redis.Port)
	redisType := common.Conf.Redis.RedisType
	MaxIdle := common.Conf.Redis.MaxIdle
	masterName := common.Conf.Redis.MasterName
	password := common.Conf.Redis.Password
	db := common.Conf.Redis.Db
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

		common.RedisConnPool = &redis.Pool{
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
		common.RedisConnPool = &redis.Pool{
			MaxIdle:     MaxIdle,
			IdleTimeout: 240 * time.Second,
			// Dial or DialContext must be set. When both are set, DialContext takes precedence over Dial.
			Dial: func() (redis.Conn, error) {
				setdb := redis.DialDatabase(db)
				pd := redis.DialPassword(password)
				c, err := redis.Dial("tcp", redisAddr, setdb, pd)
				if err != nil {
					c.Close()
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
