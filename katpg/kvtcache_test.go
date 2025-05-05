package katpg

import (
	"context"
	"errors"
	"github.com/mobiletoly/gokatana/katcache"
	"github.com/mobiletoly/gokatana/kattest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func setupTestCache(t *testing.T) (*KVTCache, context.Context) {
	ctx := kattest.AppTestContext()
	pc := RunPostgresTestContainer(ctx, t, nil, nil)
	pool := pc.BuildPgxPool(ctx, t)
	t.Cleanup(func() {
		defer pool.Close()
		pc.Terminate(ctx, t)
	})
	cache, err := NewKVTCache(ctx, pool, "public", nil)
	require.NoError(t, err)

	return cache, ctx
}

func TestKVTCache_BasicOperations(t *testing.T) {
	cache, ctx := setupTestCache(t)

	// Register a collection with string value type.
	collection := katcache.Collection{
		Name:      "test_cache",
		Ttl:       10 * time.Second,
		ValueType: katcache.CollectionValueTypeString,
	}
	cache.Register(ctx, collection)

	// Test Set and Get operations.
	key := katcache.CollectionKey{Name: "test_cache", Key: "key1"}
	value := "value1"
	err := cache.Set(ctx, key, value)
	require.NoError(t, err)

	var fetchedValue string
	found, err := cache.Get(ctx, key, &fetchedValue)
	require.NoError(t, err)
	assert.True(t, found, "Expected to find the key in the cache")
	assert.Equal(t, value, fetchedValue, "Expected fetched value to match the set value")

	// Test Del operation.
	err = cache.Del(ctx, key)
	require.NoError(t, err)
	found, err = cache.Get(ctx, key, &fetchedValue)
	require.NoError(t, err)
	assert.False(t, found, "Expected the key to be deleted from the cache")
}

func TestKVTCache_Expiration(t *testing.T) {
	cache, ctx := setupTestCache(t)

	// Register a collection with a short TTL.
	collection := katcache.Collection{
		Name:      "test_cache",
		Ttl:       2 * time.Second, // Short TTL for testing expiration.
		ValueType: katcache.CollectionValueTypeString,
	}
	cache.Register(ctx, collection)

	// Test Set and Expiration.
	key := katcache.CollectionKey{Name: "test_cache", Key: "key1"}
	value := "value1"
	err := cache.Set(ctx, key, value)
	require.NoError(t, err)

	var fetchedValue string
	found, err := cache.Get(ctx, key, &fetchedValue)
	require.NoError(t, err)
	assert.True(t, found, "Expected to find the key in the cache")
	assert.Equal(t, value, fetchedValue, "Expected fetched value to match the set value")

	// Wait for the TTL to expire.
	time.Sleep(3 * time.Second)

	found, err = cache.Get(ctx, key, &fetchedValue)
	require.NoError(t, err)
	assert.False(t, found, "Expected the key to have expired and been removed from the cache")
}

func TestKVTCache_IntValueType(t *testing.T) {
	cache, ctx := setupTestCache(t)

	// Register a collection with int64 value type.
	collection := katcache.Collection{
		Name:      "test_int_cache",
		Ttl:       10 * time.Second,
		ValueType: katcache.CollectionValueTypeInt,
	}
	cache.Register(ctx, collection)

	// Test Set and Get operations for int64 values.
	key := katcache.CollectionKey{Name: "test_int_cache", Key: "key1"}
	value := int64(42)
	err := cache.Set(ctx, key, value)
	require.NoError(t, err)

	var fetchedValue int64
	found, err := cache.Get(ctx, key, &fetchedValue)
	require.NoError(t, err)
	assert.True(t, found, "Expected to find the key in the cache")
	assert.Equal(t, value, fetchedValue, "Expected fetched value to match the set value")

	// Test Del operation.
	err = cache.Del(ctx, key)
	require.NoError(t, err)
	found, err = cache.Get(ctx, key, &fetchedValue)
	require.NoError(t, err)
	assert.False(t, found, "Expected the key to be deleted from the cache")
}

func TestKVTCache_GetOrSet(t *testing.T) {
	cache, ctx := setupTestCache(t)

	// Register a collection with string value type.
	collection := katcache.Collection{
		Name:      "test_cache",
		Ttl:       10 * time.Second,
		ValueType: katcache.CollectionValueTypeString,
	}
	cache.Register(ctx, collection)

	// Define a fetch function for fetching data if not found in the cache.
	fetchFunc := func(key katcache.CollectionKey) (string, error) {
		if key.Key == "missing_key" {
			return "", errors.New("key not found in data source")
		}
		return "fetched_value", nil
	}

	// Test GetOrSetAsync with a missing key.
	key := katcache.CollectionKey{Name: "test_cache", Key: "key1"}
	value, err := katcache.GetOrSetAsync[string](ctx, cache, key, fetchFunc)
	time.Sleep(1 * time.Second) // Wait for the value to be set in the cache.
	require.NoError(t, err)
	assert.Equal(t, "fetched_value", value, "Expected fetched value to be set in the cache")

	// Verify the value is now stored in the cache.
	var cachedValue string
	found, err := cache.Get(ctx, key, &cachedValue)
	require.NoError(t, err)
	assert.True(t, found, "Expected the key to be found in the cache")
	assert.Equal(t, "fetched_value", cachedValue, "Expected cached value to match fetched value")

	// Test GetOrSetAsync with an existing key.
	value, err = katcache.GetOrSetAsync[string](ctx, cache, key, fetchFunc)
	require.NoError(t, err)
	assert.Equal(t, "fetched_value", value, "Expected to retrieve the existing cached value")

	// Test GetOrSetAsync with a key causing an error in fetch.
	errorKey := katcache.CollectionKey{Name: "test_cache", Key: "missing_key"}
	_, err = katcache.GetOrSetAsync[string](ctx, cache, errorKey, fetchFunc)
	assert.Error(t, err, "Expected an error when fetching data for a missing key")
	assert.EqualError(t, err, "key not found in data source")
}

func TestKVTCache_GetOrSet_JSONObject(t *testing.T) {
	cache, ctx := setupTestCache(t)

	// Register a collection for JSON objects.
	collection := katcache.Collection{
		Name:      "test_json_object_cache",
		Ttl:       10 * time.Second,
		ValueType: katcache.CollectionValueTypeObject,
	}
	cache.Register(ctx, collection)

	// Define a fetch function for JSON objects.
	type TestData struct {
		ID    int    `json:"id"`
		Name  string `json:"name"`
		Email string `json:"email"`
	}
	fetchFunc := func(key katcache.CollectionKey) (TestData, error) {
		if key.Key == "missing_key" {
			return TestData{}, errors.New("key not found in data source")
		}
		return TestData{ID: 42, Name: "Alice", Email: "alice@example.com"}, nil
	}

	// Test GetOrSetAsync with a missing key.
	key := katcache.CollectionKey{Name: "test_json_object_cache", Key: "key1"}
	value, err := katcache.GetOrSetAsync[TestData](ctx, cache, key, fetchFunc)
	time.Sleep(1 * time.Second) // Allow time for caching.
	require.NoError(t, err)
	expectedValue := TestData{ID: 42, Name: "Alice", Email: "alice@example.com"}
	assert.Equal(t, expectedValue, value, "Expected fetched JSON object to be set in the cache")

	// Verify the value is now stored in the cache.
	var cachedValue TestData
	found, err := cache.Get(ctx, key, &cachedValue)
	require.NoError(t, err)
	assert.True(t, found, "Expected the key to be found in the cache")
	assert.Equal(t, expectedValue, cachedValue, "Expected cached JSON object to match fetched value")

	// Test GetOrSetAsync with an existing key.
	value, err = katcache.GetOrSetAsync[TestData](ctx, cache, key, fetchFunc)
	require.NoError(t, err)
	assert.Equal(t, expectedValue, value, "Expected to retrieve the existing cached JSON object")

	// Test GetOrSetAsync with a key causing an error in fetch.
	errorKey := katcache.CollectionKey{Name: "test_json_object_cache", Key: "missing_key"}
	_, err = katcache.GetOrSetAsync[TestData](ctx, cache, errorKey, fetchFunc)
	assert.Error(t, err, "Expected an error when fetching data for a missing key")
	assert.EqualError(t, err, "key not found in data source")
}
