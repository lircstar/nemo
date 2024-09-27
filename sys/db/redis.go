package db

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/lircstar/nemo/sys/log"
	"time"
)

// RedisDB represents a Redis database connection.
type RedisDB struct {
	Address  string
	Password string
	DB       int
	client   *redis.Client
	ctx      context.Context
}

// NewRedisDB creates a new RedisDB instance and initializes the connection.
func NewRedisDB(address string, password string, db int) *RedisDB {
	rd := &RedisDB{
		Address:  address,
		Password: password,
		DB:       db,
	}
	rd.client = rd.CreateClient()

	// Check connection
	if !rd.checkConnection() {
		log.Fatalf("Failed to connect to Redis: %v", rd.client.Ping(rd.ctx).Err())
	}
	return rd
}

// CreateClient initializes the Redis client.
func (r *RedisDB) CreateClient() *redis.Client {
	r.ctx = context.Background()
	r.client = redis.NewClient(&redis.Options{
		Addr:     r.Address,
		Password: r.Password,
		DB:       r.DB,
	})
	return r.client
}

// checkConnection checks if the Redis connection is successful.
func (r *RedisDB) checkConnection() bool {
	startTime := time.Now().Unix()
	for {
		err := r.client.Ping(r.ctx).Err()
		if err == nil {
			return true
		}
		if time.Now().Unix()-startTime > 10 {
			return false
		}
		time.Sleep(2 * time.Second)
	}
}

// SetNx sets a key-value pair in Redis if it does not exist.
func (r *RedisDB) SetNx(key string, value any) bool {
	val, err := r.client.SetNX(r.ctx, key, value, 0).Result()
	if err != nil {
		fmt.Println("Error setting value:", err)
		return false
	}
	return val
}

// SetEx sets a key-value pair in Redis with an expiration time.
func (r *RedisDB) SetEx(key string, value any, expiration time.Duration) bool {
	err := r.client.Set(r.ctx, key, value, expiration).Err()
	if err != nil {
		fmt.Println("Error setting value:", err)
		return false
	}
	return true
}

// Set sets a key-value pair in Redis.
func (r *RedisDB) Set(key string, value any) bool {
	err := r.client.Set(r.ctx, key, value, 0).Err()
	if err != nil {
		fmt.Println("Error setting value:", err)
		return false
	}
	return true
}

// Get retrieves the value of a key from Redis.
func (r *RedisDB) Get(key string) string {
	val, err := r.client.Get(r.ctx, key).Result()
	if err != nil {
		fmt.Println("Error getting value:", err)
		return ""
	}
	return val
}

// GetKeys retrieves all keys from Redis.
func (r *RedisDB) GetKeys() ([]string, error) {
	keys, err := r.client.Keys(r.ctx, "*").Result()
	if err != nil {
		return nil, err
	}
	return keys, nil
}

// Del deletes a key from Redis.
func (r *RedisDB) Del(key string) error {
	return r.client.Del(r.ctx, key).Err()
}

// LPush pushes a value to a list in Redis.
func (r *RedisDB) LPush(key string, values ...any) error {
	return r.client.LPush(r.ctx, key, values).Err()
}

// LPop pops a value from a list in Redis.
func (r *RedisDB) LPop(key string) (string, error) {
	val, err := r.client.LPop(r.ctx, key).Result()
	if err != nil {
		return "", err
	}
	return val, nil
}

// RPush pushes a value to a list in Redis.
func (r *RedisDB) RPush(key string, values ...any) error {
	return r.client.RPush(r.ctx, key, values).Err()
}

// RPop pops a value from a list in Redis.
func (r *RedisDB) RPop(key string) (string, error) {
	val, err := r.client.RPop(r.ctx, key).Result()
	if err != nil {
		return "", err
	}
	return val, nil
}

// LRange retrieves a range of values from a list in Redis.
func (r *RedisDB) LRange(key string, start, stop int64) ([]string, error) {
	val, err := r.client.LRange(r.ctx, key, start, stop).Result()
	if err != nil {
		return nil, err
	}
	return val, nil
}

// LLen retrieves the length of a list in Redis.
func (r *RedisDB) LLen(key string) (int64, error) {
	val, err := r.client.LLen(r.ctx, key).Result()
	if err != nil {
		return 0, err
	}
	return val, nil
}

// LRem removes a value from a list in Redis.
func (r *RedisDB) LRem(key string, count int64, value any) (int64, error) {
	val, err := r.client.LRem(r.ctx, key, count, value).Result()
	if err != nil {
		return 0, err
	}
	return val, nil
}

// LSet sets a value in a list in Redis.
func (r *RedisDB) LSet(key string, index int64, value any) error {
	return r.client.LSet(r.ctx, key, index, value).Err()
}

// LTrim trims a list in Redis.
func (r *RedisDB) LTrim(key string, start, stop int64) error {
	return r.client.LTrim(r.ctx, key, start, stop).Err()
}

// LIndex retrieves a value from a list by index in Redis.
func (r *RedisDB) LIndex(key string, index int64) (string, error) {
	val, err := r.client.LIndex(r.ctx, key, index).Result()
	if err != nil {
		return "", err
	}
	return val, nil
}

// LInsert inserts a value to a list in Redis.
func (r *RedisDB) LInsert(key, op, pivot, value string) (int64, error) {
	val, err := r.client.LInsert(r.ctx, key, op, pivot, value).Result()
	if err != nil {
		return 0, err
	}
	return val, nil
}

// LPopCount pops multiple values from a list in Redis.
func (r *RedisDB) LPopCount(key string, count int) ([]string, error) {
	val, err := r.client.LPopCount(r.ctx, key, count).Result()
	if err != nil {
		return nil, err
	}
	return val, nil
}

// RPopCount pops multiple values from a list in Redis.
func (r *RedisDB) RPopCount(key string, count int) ([]string, error) {
	val, err := r.client.RPopCount(r.ctx, key, count).Result()
	if err != nil {
		return nil, err
	}
	return val, nil
}

// RPopLPush pops a value from a list and pushes it to another list in Redis.
func (r *RedisDB) RPopLPush(source, destination string) (string, error) {
	val, err := r.client.RPopLPush(r.ctx, source, destination).Result()
	if err != nil {
		return "", err
	}
	return val, nil
}

// SAdd adds a member to a set in Redis.
func (r *RedisDB) SAdd(key string, members ...any) error {
	return r.client.SAdd(r.ctx, key, members).Err()
}

// SCard retrieves the cardinality of a set in Redis.
func (r *RedisDB) SCard(key string) (int64, error) {
	val, err := r.client.SCard(r.ctx, key).Result()
	if err != nil {
		return 0, err
	}
	return val, nil
}

// SDiff retrieves the difference between sets in Redis.
func (r *RedisDB) SDiff(keys ...string) ([]string, error) {
	val, err := r.client.SDiff(r.ctx, keys...).Result()
	if err != nil {
		return nil, err
	}
	return val, nil
}

// SInter retrieves the intersection between sets in Redis.
func (r *RedisDB) SInter(keys ...string) ([]string, error) {
	val, err := r.client.SInter(r.ctx, keys...).Result()
	if err != nil {
		return nil, err
	}
	return val, nil
}

// SIsMember checks if a member exists in a set in Redis.
func (r *RedisDB) SIsMember(key string, member any) (bool, error) {
	val, err := r.client.SIsMember(r.ctx, key, member).Result()
	if err != nil {
		return false, err
	}
	return val, nil
}

// SMembers retrieves all members of a set in Redis.
func (r *RedisDB) SMembers(key string) ([]string, error) {
	val, err := r.client.SMembers(r.ctx, key).Result()
	if err != nil {
		return nil, err
	}
	return val, nil
}

// SMove moves a member from one set to another in Redis.
func (r *RedisDB) SMove(source, destination string, member any) (bool, error) {
	val, err := r.client.SMove(r.ctx, source, destination, member).Result()
	if err != nil {
		return false, err
	}
	return val, nil
}

// SPop pops a member from a set in Redis.
func (r *RedisDB) SPop(key string) (string, error) {
	val, err := r.client.SPop(r.ctx, key).Result()
	if err != nil {
		return "", err
	}
	return val, nil
}

// SRandMember retrieves a random member from a set in Redis.
func (r *RedisDB) SRandMember(key string) (string, error) {
	val, err := r.client.SRandMember(r.ctx, key).Result()
	if err != nil {
		return "", err
	}
	return val, nil
}

// SRem removes a member from a set in Redis.
func (r *RedisDB) SRem(key string, members ...any) error {
	return r.client.SRem(r.ctx, key, members).Err()
}

// SUnion retrieves the union between sets in Redis.
func (r *RedisDB) SUnion(keys ...string) ([]string, error) {
	val, err := r.client.SUnion(r.ctx, keys...).Result()
	if err != nil {
		return nil, err
	}
	return val, nil
}

// ZAdd adds a member to a sorted set in Redis.
func (r *RedisDB) ZAdd(key string, members ...*redis.Z) error {
	return r.client.ZAdd(r.ctx, key, members...).Err()
}

// ZCard retrieves the cardinality of a sorted set in Redis.
func (r *RedisDB) ZCard(key string) (int64, error) {
	val, err := r.client.ZCard(r.ctx, key).Result()
	if err != nil {
		return 0, err
	}
	return val, nil
}

// ZCount retrieves the count of members in a sorted set within a score range in Redis.
func (r *RedisDB) ZCount(key, min, max string) (int64, error) {
	val, err := r.client.ZCount(r.ctx, key, min, max).Result()
	if err != nil {
		return 0, err
	}
	return val, nil
}

// ZIncrBy increments the score of a member in a sorted set in Redis.
func (r *RedisDB) ZIncrBy(key string, increment float64, member string) (float64, error) {
	val, err := r.client.ZIncrBy(r.ctx, key, increment, member).Result()
	if err != nil {
		return 0, err
	}
	return val, nil
}

// ZRange retrieves a range of members from a sorted set in Redis.
func (r *RedisDB) ZRange(key string, start, stop int64) ([]string, error) {
	val, err := r.client.ZRange(r.ctx, key, start, stop).Result()
	if err != nil {
		return nil, err
	}
	return val, nil
}

// ZRangeWithScores retrieves a range of members with scores from a sorted set in Redis.
func (r *RedisDB) ZRangeWithScores(key string, start, stop int64) ([]redis.Z, error) {
	val, err := r.client.ZRangeWithScores(r.ctx, key, start, stop).Result()
	if err != nil {
		return nil, err
	}
	return val, nil
}

// ZRangeByScore retrieves a range of members with scores from a sorted set within a score range in Redis.
func (r *RedisDB) ZRangeByScore(key string, opt *redis.ZRangeBy) ([]string, error) {
	val, err := r.client.ZRangeByScore(r.ctx, key, opt).Result()
	if err != nil {
		return nil, err
	}
	return val, nil
}

// ZRangeByScoreWithScores retrieves a range of members with scores from a sorted set within a score range in Redis.
func (r *RedisDB) ZRangeByScoreWithScores(key string, opt *redis.ZRangeBy) ([]redis.Z, error) {
	val, err := r.client.ZRangeByScoreWithScores(r.ctx, key, opt).Result()
	if err != nil {
		return nil, err
	}
	return val, nil
}

// ZRank retrieves the rank of a member in a sorted set in Redis.
func (r *RedisDB) ZRank(key, member string) (int64, error) {
	val, err := r.client.ZRank(r.ctx, key, member).Result()
	if err != nil {
		return 0, err
	}
	return val, nil
}

// ZRem removes members from a sorted set in Redis.
func (r *RedisDB) ZRem(key string, members ...any) error {
	return r.client.ZRem(r.ctx, key, members).Err()
}

// ZRemRangeByRank removes members from a sorted set by rank in Redis.
func (r *RedisDB) ZRemRangeByRank(key string, start, stop int64) (int64, error) {
	val, err := r.client.ZRemRangeByRank(r.ctx, key, start, stop).Result()
	if err != nil {
		return 0, err
	}
	return val, nil
}

// ZRemRangeByScore removes members from a sorted set by score in Redis.
func (r *RedisDB) ZRemRangeByScore(key, min, max string) (int64, error) {
	val, err := r.client.ZRemRangeByScore(r.ctx, key, min, max).Result()
	if err != nil {
		return 0, err
	}
	return val, nil
}

// ZRevRange retrieves a range of members from a sorted set in reverse order in Redis.
func (r *RedisDB) ZRevRange(key string, start, stop int64) ([]string, error) {
	val, err := r.client.ZRevRange(r.ctx, key, start, stop).Result()
	if err != nil {
		return nil, err
	}
	return val, nil
}

// ZRevRangeWithScores retrieves a range of members with scores from a sorted set in reverse order in Redis.
func (r *RedisDB) ZRevRangeWithScores(key string, start, stop int64) ([]redis.Z, error) {
	val, err := r.client.ZRevRangeWithScores(r.ctx, key, start, stop).Result()
	if err != nil {
		return nil, err
	}
	return val, nil
}

// ZRevRangeByScore retrieves a range of members with scores from a sorted set within a score range in reverse order in Redis.
func (r *RedisDB) ZRevRangeByScore(key string, opt *redis.ZRangeBy) ([]string, error) {
	val, err := r.client.ZRevRangeByScore(r.ctx, key, opt).Result()
	if err != nil {
		return nil, err
	}
	return val, nil
}

// ZRevRangeByScoreWithScores retrieves a range of members with scores from a sorted set within a score range in reverse order in Redis.
func (r *RedisDB) ZRevRangeByScoreWithScores(key string, opt *redis.ZRangeBy) ([]redis.Z, error) {
	val, err := r.client.ZRevRangeByScoreWithScores(r.ctx, key, opt).Result()
	if err != nil {
		return nil, err
	}
	return val, nil
}

// ZRevRank retrieves the rank of a member in a sorted set in reverse order in Redis.
func (r *RedisDB) ZRevRank(key, member string) (int64, error) {
	val, err := r.client.ZRevRank(r.ctx, key, member).Result()
	if err != nil {
		return 0, err
	}
	return val, nil
}

// ZScore retrieves the score of a member in a sorted set in Redis.
func (r *RedisDB) ZScore(key, member string) (float64, error) {
	val, err := r.client.ZScore(r.ctx, key, member).Result()
	if err != nil {
		return 0, err
	}
	return val, nil
}

// HIncrBy increments a field in a hash in Redis.
func (r *RedisDB) HIncrBy(key, field string, value int64) (int64, error) {
	val, err := r.client.HIncrBy(r.ctx, key, field, value).Result()
	if err != nil {
		return 0, err
	}
	return val, nil
}

// HIncrByFloat increments a field in a hash in Redis by a float value.
func (r *RedisDB) HIncrByFloat(key, field string, value float64) (float64, error) {
	val, err := r.client.HIncrByFloat(r.ctx, key, field, value).Result()
	if err != nil {
		return 0, err
	}
	return val, nil
}

// HKeys retrieves all fields from a hash in Redis.
func (r *RedisDB) HKeys(key string) ([]string, error) {
	keys, err := r.client.HKeys(r.ctx, key).Result()
	if err != nil {
		return nil, err
	}
	return keys, nil
}

// HSetNx sets a field in a hash in Redis if it does not exist.
func (r *RedisDB) HSetNx(key, field string, value any) bool {
	val, err := r.client.HSetNX(r.ctx, key, field, value).Result()
	if err != nil {
		fmt.Println("Error setting value:", err)
		return false
	}
	return val
}

// HExists checks if a field exists in a hash in Redis.
func (r *RedisDB) HExists(key, field string) bool {
	val, err := r.client.HExists(r.ctx, key, field).Result()
	if err != nil {
		fmt.Println("Error checking value:", err)
		return false
	}
	return val
}

// HSet sets a field in a hash in Redis.
func (r *RedisDB) HSet(key, field string, value any) bool {
	err := r.client.HSet(r.ctx, key, field, value).Err()
	if err != nil {
		fmt.Println("Error setting value:", err)
		return false
	}
	return true
}

// HGet retrieves a field from a hash in Redis.
func (r *RedisDB) HGet(key, field string) string {
	val, err := r.client.HGet(r.ctx, key, field).Result()
	if err != nil {
		fmt.Println("Error getting value:", err)
		return ""
	}
	return val
}

// HGetAll retrieves all fields from a hash in Redis.
func (r *RedisDB) HGetAll(key string) (map[string]string, error) {
	val, err := r.client.HGetAll(r.ctx, key).Result()
	if err != nil {
		return nil, err
	}
	return val, nil
}

// HDel deletes a field from a hash in Redis.
func (r *RedisDB) HDel(key, field string) error {
	return r.client.HDel(r.ctx, key, field).Err()
}

// HmSet sets multiple fields in a hash in Redis.
func (r *RedisDB) HmSet(key string, fields map[string]any) bool {
	err := r.client.HMSet(r.ctx, key, fields).Err()
	if err != nil {
		fmt.Println("Error setting value:", err)
		return false
	}
	return true
}

// HmGet retrieves multiple fields from a hash in Redis.
func (r *RedisDB) HmGet(key string, fields ...string) ([]any, error) {
	val, err := r.client.HMGet(r.ctx, key, fields...).Result()
	if err != nil {
		return nil, err
	}
	return val, nil
}

// Close closes the Redis client connection.
func (r *RedisDB) Close() {
	err := r.client.Close()
	if err != nil {
		return
	}
}
