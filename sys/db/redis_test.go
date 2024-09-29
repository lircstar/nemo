package db

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var testRedisDB *RedisDB

func setup() {
	testRedisDB = NewRedisDB("localhost:6379", "", 0)
}

func teardown() {
	testRedisDB.Close()
}

func TestNewRedisDB(t *testing.T) {
	setup()
	defer teardown()

	assert.NotNil(t, testRedisDB.GetClient(), "Redis client should be initialized")
}

func TestSetAndGet(t *testing.T) {
	setup()
	defer teardown()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	key := "test_key"
	value := "test_value"

	err := testRedisDB.GetClient().Set(ctx, key, value, 0).Err()
	assert.NoError(t, err, "Set operation should succeed")

	result, err := testRedisDB.GetClient().Get(ctx, key).Result()
	assert.NoError(t, err, "Get operation should succeed")
	assert.Equal(t, value, result, "Get operation should return the correct value")
}

func TestGetKeys(t *testing.T) {
	setup()
	defer teardown()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	key := "test_key"
	value := "test_value"
	err := testRedisDB.GetClient().Set(ctx, key, value, 0).Err()
	assert.NoError(t, err, "Set operation should succeed")

	keys, err := testRedisDB.GetClient().Keys(ctx, "*").Result()
	assert.NoError(t, err, "GetKeys operation should not return an error")
	assert.Contains(t, keys, key, "GetKeys should return the set key")
}

func TestDel(t *testing.T) {
	setup()
	defer teardown()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	key := "test_key"
	value := "test_value"
	err := testRedisDB.GetClient().Set(ctx, key, value, 0).Err()
	assert.NoError(t, err, "Set operation should succeed")

	err = testRedisDB.GetClient().Del(ctx, key).Err()
	assert.NoError(t, err, "Del operation should not return an error")

	result, err := testRedisDB.GetClient().Get(ctx, key).Result()
	assert.Error(t, err, "Get operation should return an error for deleted key")
	assert.Empty(t, result, "Get operation should return an empty string for deleted key")
}

func TestHSetAndHGet(t *testing.T) {
	setup()
	defer teardown()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	key := "test_hash"
	field := "field1"
	value := "value1"

	err := testRedisDB.GetClient().HSet(ctx, key, field, value).Err()
	assert.NoError(t, err, "HSet operation should succeed")

	result, err := testRedisDB.GetClient().HGet(ctx, key, field).Result()
	assert.NoError(t, err, "HGet operation should succeed")
	assert.Equal(t, value, result, "HGet operation should return the correct value")
}

func TestHGetAll(t *testing.T) {
	setup()
	defer teardown()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	key := "test_hash"
	fields := map[string]interface{}{"field1": "value1", "field2": "value2"}
	err := testRedisDB.GetClient().HMSet(ctx, key, fields).Err()
	assert.NoError(t, err, "HMSet operation should succeed")

	result, err := testRedisDB.GetClient().HGetAll(ctx, key).Result()
	assert.NoError(t, err, "HGetAll operation should not return an error")
	assert.Equal(t, "value1", result["field1"], "HGetAll should return the correct value for field1")
	assert.Equal(t, "value2", result["field2"], "HGetAll should return the correct value for field2")
}

func TestHmSetAndHmGet(t *testing.T) {
	setup()
	defer teardown()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	key := "test_hash"
	fields := map[string]interface{}{"field1": "value1", "field2": "value2"}

	err := testRedisDB.GetClient().HMSet(ctx, key, fields).Err()
	assert.NoError(t, err, "HMSet operation should succeed")

	result, err := testRedisDB.GetClient().HMGet(ctx, key, "field1", "field2").Result()
	assert.NoError(t, err, "HMGet operation should not return an error")
	assert.Equal(t, "value1", result[0], "HMGet should return the correct value for field1")
	assert.Equal(t, "value2", result[1], "HMGet should return the correct value for field2")
}
