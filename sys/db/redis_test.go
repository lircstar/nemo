package db

import (
	"context"
	_ "github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"testing"
	_ "time"
)

var testRedisDB *RedisDB

func setup() {
	testRedisDB = &RedisDB{
		Address:  "localhost:6379",
		Password: "",
		DB:       0,
		ctx:      context.Background(),
	}
	testRedisDB.client = testRedisDB.CreateClient()
}

func teardown() {
	testRedisDB.Close()
}

func TestNewRedisDB(t *testing.T) {
	setup()
	defer teardown()

	assert.NotNil(t, testRedisDB.client, "Redis client should be initialized")
}

func TestSetAndGet(t *testing.T) {
	setup()
	defer teardown()

	key := "test_key"
	value := "test_value"

	success := testRedisDB.Set(key, value)
	assert.True(t, success, "Set operation should succeed")

	result := testRedisDB.Get(key)
	assert.Equal(t, value, result, "Get operation should return the correct value")
}

func TestGetKeys(t *testing.T) {
	setup()
	defer teardown()

	key := "test_key"
	value := "test_value"
	testRedisDB.Set(key, value)

	keys, err := testRedisDB.GetKeys()
	assert.NoError(t, err, "GetKeys operation should not return an error")
	assert.Contains(t, keys, key, "GetKeys should return the set key")
}

func TestDel(t *testing.T) {
	setup()
	defer teardown()

	key := "test_key"
	value := "test_value"
	testRedisDB.Set(key, value)

	err := testRedisDB.Del(key)
	assert.NoError(t, err, "Del operation should not return an error")

	result := testRedisDB.Get(key)
	assert.Empty(t, result, "Get operation should return an empty string for deleted key")
}

func TestHSetAndHGet(t *testing.T) {
	setup()
	defer teardown()

	key := "test_hash"
	field := "field1"
	value := "value1"

	success := testRedisDB.HSet(key, field, value)
	assert.True(t, success, "HSet operation should succeed")

	result := testRedisDB.HGet(key, field)
	assert.Equal(t, value, result, "HGet operation should return the correct value")
}

func TestHGetAll(t *testing.T) {
	setup()
	defer teardown()

	key := "test_hash"
	fields := map[string]any{"field1": "value1", "field2": "value2"}
	testRedisDB.HmSet(key, fields)

	result, err := testRedisDB.HGetAll(key)
	assert.NoError(t, err, "HGetAll operation should not return an error")
	assert.Equal(t, "value1", result["field1"], "HGetAll should return the correct value for field1")
	assert.Equal(t, "value2", result["field2"], "HGetAll should return the correct value for field2")
}

func TestHmSetAndHmGet(t *testing.T) {
	setup()
	defer teardown()

	key := "test_hash"
	fields := map[string]any{"field1": "value1", "field2": "value2"}

	success := testRedisDB.HmSet(key, fields)
	assert.True(t, success, "HmSet operation should succeed")

	result, err := testRedisDB.HmGet(key, "field1", "field2")
	assert.NoError(t, err, "HmGet operation should not return an error")
	assert.Equal(t, "value1", result[0], "HmGet should return the correct value for field1")
	assert.Equal(t, "value2", result[1], "HmGet should return the correct value for field2")
}
