package redis

import (
	"context"
	"github.com/redis/go-redis/v9"
	"time"
)

type Config struct {
	Host     string
	Password string
	DB       int
}

type Client struct {
	client *redis.Client
}

func NewClient(cfg *Config) *Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Host,
		Password: cfg.Password, // no password set
		DB:       cfg.DB,       // use default DB
	})

	return &Client{client: rdb}
}

func (c *Client) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err := c.client.Ping(ctx).Result()
	return err
}
