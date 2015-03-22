package store

import (
	"os"
	"strings"
	"time"

	//"code.google.com/p/go-uuid/uuid"
	"github.com/garyburd/redigo/redis"
)

var (
	DefaultRedisUrlEnv = "REDIS_URL"
)

type RedisStore struct {
	pool *redis.Pool
}

type RedisConfig struct {
	Host, Port, User, Pass string
	Db                     int
}

func NewRedisStore() Store {
	return &RedisStore{pool: NewRedisPool(NewRedisConfig())}
}

func (*RedisStore) Read(i Item) error {
	return nil
}

func (*RedisStore) Write(i Item) error {
	return nil
}

func NewRedisConfig() *RedisConfig {
	// Use local Redis instance by default
	c := &RedisConfig{
		Host: "127.0.0.1",
		Port: "6379",
	}

	// Override default configure when env var $REDIS_URL
	// is present with the format redis://user:pass@host:port
	if url := os.Getenv(DefaultRedisUrlEnv); len(url) > 0 {
		url = strings.TrimPrefix(url, "redis://")
		parts := strings.Split(url, "@")

		// username and/or password exists
		if len(parts) == 2 {
			auth := strings.Split(parts[0], ":")
			c.User, c.Pass = auth[0], auth[1]
		}

		// the last part of the url is the address
		if addr := parts[len(parts)-1]; len(addr) > 0 {
			addrparts := strings.Split(addr, ":")
			c.Host, c.Port = addrparts[0], addrparts[1]
		}
	}
	return c
}

// TODO: Implement auth
func NewRedisPool(config *RedisConfig) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", config.Host+":"+config.Port)
			if err != nil {
				return nil, err
			}
			if len(config.Pass) > 0 {
				if _, err := c.Do("AUTH", config.Pass); err != nil {
					c.Close()
					return nil, err
				}
			}
			if config.Db > 0 {
				if _, err := c.Do("SELECT", config.Db); err != nil {
					c.Close()
					return nil, err
				}
			}
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
}
