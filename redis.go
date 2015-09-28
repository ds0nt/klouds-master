package main

import (
	"time"

	"github.com/garyburd/redigo/redis"
)

func newPool() *redis.Pool {
	return &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", Config.RedisServer)
			if err != nil {
				return nil, err
			}
			if _, err := c.Do("AUTH", Config.RedisPassword); err != nil {
				c.Close()
				return nil, err
			}
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
}

var (
	pool *redis.Pool
)
