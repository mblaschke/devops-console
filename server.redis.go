package main

import (
	"context"
	"encoding/json"
	"time"
)

func (c *Server) redisCacheSet(key string, data interface{}, expiry time.Duration) error {
	ctx := context.Background()

	if val, err := json.Marshal(data); err == nil {
		c.redis.Set(ctx, key, val, expiry)
	} else {
		return err
	}

	return nil
}

func (c *Server) redisCacheLoad(key string, data interface{}) bool {
	ctx := context.Background()

	val, err := c.redis.Get(ctx, key).Result()
	if err == nil {
		if err := json.Unmarshal([]byte(val), data); err == nil {
			return true
		}
	}

	return false
}

func (c *Server) redisCacheInvalidate(key string) {
	ctx := context.Background()
	c.redis.Expire(ctx, key, 0)
}
