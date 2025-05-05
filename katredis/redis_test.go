package katredis

import (
	"errors"
	"github.com/mobiletoly/gokatana/katcache"
	"github.com/mobiletoly/gokatana/kattest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestRedisCache_BasicOperations(t *testing.T) {
	// 1. Spin up a Redis container.
	ctx := kattest.AppTestContext()
	rc := RunRedisTestContainer(ctx, t)
	t.Cleanup(func() {
		rc.Terminate(ctx, t)
	})

	// 2. Create a RedisCache instance.
	cache, err := NewRedisCache(ctx, rc.Address(), 0)
	require.NoError(t, err)
	defer cache.Close()

	// 3. Register a collection with string value type.
	collection := katcache.Collection{
		Name:      "test_cache",
		Ttl:       10 * time.Second,
		ValueType: katcache.CollectionValueTypeString,
	}
	cache.Register(ctx, collection)

	// 4. Test Set and Get operations.
	key := katcache.CollectionKey{Name: "test_cache", Key: "key1"}
	value := "value1"
	err = cache.Set(ctx, key, value)
	require.NoError(t, err)

	var fetchedValue string
	found, err := cache.Get(ctx, key, &fetchedValue)
	require.NoError(t, err)
	assert.True(t, found, "Expected to find the key in the cache")
	assert.Equal(t, value, fetchedValue, "Expected fetched value to match the set value")

	// 5. Test Del operation.
	err = cache.Del(ctx, key)
	require.NoError(t, err)

	found, err = cache.Get(ctx, key, &fetchedValue)
	require.NoError(t, err)
	assert.False(t, found, "Expected the key to be deleted from the cache")
}

func TestRedisCache_Expiration(t *testing.T) {
	// 1. Spin up a Redis container.
	ctx := kattest.AppTestContext()
	rc := RunRedisTestContainer(ctx, t)
	t.Cleanup(func() {
		rc.Terminate(ctx, t)
	})

	// 2. Create a RedisCache instance.
	cache, err := NewRedisCache(ctx, rc.Address(), 0)
	require.NoError(t, err)
	defer cache.Close()

	// 3. Register a collection with a short TTL.
	collection := katcache.Collection{
		Name:      "test_cache",
		Ttl:       2 * time.Second, // Short TTL for testing expiration.
		ValueType: katcache.CollectionValueTypeString,
	}
	cache.Register(ctx, collection)

	// 4. Test Set and Expiration.
	key := katcache.CollectionKey{Name: "test_cache", Key: "key1"}
	value := "value1"
	err = cache.Set(ctx, key, value)
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

func TestRedisCache_GetOrSetAsync(t *testing.T) {
	// 1. Spin up a Redis container.
	ctx := kattest.AppTestContext()
	rc := RunRedisTestContainer(ctx, t)
	t.Cleanup(func() {
		rc.Terminate(ctx, t)
	})

	// 2. Create a RedisCache instance.
	cache, err := NewRedisCache(ctx, rc.Address(), 0)
	require.NoError(t, err)
	defer cache.Close()

	// 3. Register a collection with string value type.
	collection := katcache.Collection{
		Name:      "test_cache",
		Ttl:       10 * time.Second,
		ValueType: katcache.CollectionValueTypeString,
	}
	cache.Register(ctx, collection)

	// 4. Define a fetch function for fetching data if not found in the cache.
	fetchFunc := func(key katcache.CollectionKey) (string, error) {
		if key.Key == "missing_key" {
			return "", errors.New("key not found in data source")
		}
		return "fetched_value", nil
	}

	// 5. Test GetOrSetAsync with a missing key.
	key := katcache.CollectionKey{Name: "test_cache", Key: "key1"}
	value, err := katcache.GetOrSetAsync(ctx, cache, key, fetchFunc)
	require.NoError(t, err)
	assert.Equal(t, "fetched_value", value, "Expected fetched value to be set in the cache")
	time.Sleep(1 * time.Second)

	// Verify the value is now stored in the cache.
	var cachedValue string
	found, err := cache.Get(ctx, key, &cachedValue)
	require.NoError(t, err)
	assert.True(t, found, "Expected the key to be found in the cache")
	assert.Equal(t, "fetched_value", cachedValue, "Expected cached value to match fetched value")

	// 6. Test GetOrSetAsync with an existing key.
	value, err = katcache.GetOrSetAsync(ctx, cache, key, fetchFunc)
	require.NoError(t, err)
	assert.Equal(t, "fetched_value", value, "Expected to retrieve the existing cached value")

	// 7. Test GetOrSetAsync with a key causing an error in fetch.
	errorKey := katcache.CollectionKey{Name: "test_cache", Key: "missing_key"}
	_, err = katcache.GetOrSetAsync(ctx, cache, errorKey, fetchFunc)
	assert.Error(t, err, "Expected an error when fetching data for a missing key")
	assert.EqualError(t, err, "key not found in data source")
}
