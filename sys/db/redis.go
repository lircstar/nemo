package db

import (
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/lircstar/nemo/sys/log"
	"time"
)

// RedisDB represents a Redis database connection.
type RedisDB struct {
	address  string
	password string
	db       int
	client   *redis.Client
}

// NewRedisDB creates a new RedisDB instance and initializes the connection.
func NewRedisDB(address string, password string, db int) *RedisDB {
	rd := &RedisDB{
		address:  address,
		password: password,
		db:       db,
	}

	rd.client = redis.NewClient(&redis.Options{
		Addr:     address,
		Password: password,
		DB:       db,
	})

	// Check connection
	if err := rd.checkConnection(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	return rd
}

// checkConnection checks if the Redis connection is successful.
func (r *RedisDB) checkConnection() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	for {
		err := r.client.Ping(ctx).Err()
		if err == nil {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			time.Sleep(2 * time.Second)
		}
	}
}

func (r *RedisDB) GetDB() int {
	return r.db
}

// GetClient returns the Redis client.
func (r *RedisDB) GetClient() *redis.Client {
	return r.client
}

// Close closes the Redis client connection.
func (r *RedisDB) Close() {
	if err := r.client.Close(); err != nil {
		log.Errorf("Failed to close Redis connection: %v", err)
	}
}
