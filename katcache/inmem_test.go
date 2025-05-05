package katcache

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestInMemCacheAdapter_Register(t *testing.T) {
	cache := NewInMem()

	coll := Collection{
		Name:          "testCollection",
		ValueType:     CollectionValueTypeString,
		LocalMaxItems: 10,
		Ttl:           0,
	}

	// Register the collection
	assert.NotPanics(t, func() {
		cache.Register(context.Background(), coll)
	})

	// Registering the same collection again should panic
	assert.Panics(t, func() {
		cache.Register(context.Background(), coll)
	})
}

func TestInMemCacheAdapter_SetAndGet_String(t *testing.T) {
	cache := NewInMem()
	coll := Collection{
		Name:          "testCollection",
		ValueType:     CollectionValueTypeString,
		LocalMaxItems: 10,
		Ttl:           0,
	}
	cache.Register(context.Background(), coll)

	key := CollectionKey{
		Name: "testCollection",
		Key:  "testKey",
	}
	value := "testValue"

	assert.NotPanics(t, func() {
		cache.Set(context.Background(), key, value)
	})
	var retrievedValue string
	found, err := cache.Get(context.Background(), key, &retrievedValue)
	assert.NoError(t, err)

	assert.True(t, found)
	assert.Equal(t, value, retrievedValue)
}

func TestInMemCacheAdapter_SetAndGet_Object(t *testing.T) {
	cache := NewInMem()

	coll := Collection{
		Name:          "testCollection",
		ValueType:     CollectionValueTypeObject,
		LocalMaxItems: 10,
		Ttl:           0,
	}
	cache.Register(context.Background(), coll)

	key := CollectionKey{
		Name: "testCollection",
		Key:  "testKey",
	}

	value := map[string]interface{}{
		"field1": "value1",
		"field2": 42,
	}

	// Set the value
	assert.NotPanics(t, func() {
		cache.Set(context.Background(), key, value)
	})

	// Get the value
	var retrievedValue map[string]interface{}
	found, err := cache.Get(context.Background(), key, &retrievedValue)
	assert.NoError(t, err)
	assert.True(t, found)

	// Convert back to JSON for comparison
	expectedJSON, _ := json.Marshal(value)
	retrievedJSON, _ := json.Marshal(retrievedValue)
	assert.JSONEq(t, string(expectedJSON), string(retrievedJSON))
}

func TestInMemCacheAdapter_SetAndGet_Int64(t *testing.T) {
	cache := NewInMem()

	coll := Collection{
		Name:          "testCollection",
		ValueType:     CollectionValueTypeInt,
		LocalMaxItems: 10,
		Ttl:           0,
	}
	cache.Register(context.Background(), coll)

	key := CollectionKey{
		Name: "testCollection",
		Key:  "testKey",
	}
	value := int64(1234567890)

	err := cache.Set(context.Background(), key, value)
	assert.NoError(t, err)

	// Get the value
	var retrievedValue int64
	found, err := cache.Get(context.Background(), key, &retrievedValue)
	assert.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, value, retrievedValue)
}

func TestInMemCacheAdapter_SetAndGet_Float64(t *testing.T) {
	cache := NewInMem()

	coll := Collection{
		Name:          "testCollection",
		ValueType:     CollectionValueTypeReal,
		LocalMaxItems: 10,
		Ttl:           0,
	}
	cache.Register(context.Background(), coll)

	key := CollectionKey{
		Name: "testCollection",
		Key:  "testKey",
	}
	value := 12345.6789

	// Set the value
	err := cache.Set(context.Background(), key, value)
	assert.NoError(t, err)

	// Get the value
	var retrievedValue float64
	found, err := cache.Get(context.Background(), key, &retrievedValue)
	assert.NoError(t, err)
	assert.True(t, found)
	assert.InDelta(t, value, retrievedValue, 0.00001)
}

func TestInMemCacheAdapter_Del(t *testing.T) {
	cache := NewInMem()

	coll := Collection{
		Name:          "testCollection",
		ValueType:     CollectionValueTypeString,
		LocalMaxItems: 10,
		Ttl:           0,
	}
	cache.Register(context.Background(), coll)

	key := CollectionKey{
		Name: "testCollection",
		Key:  "testKey",
	}

	value := "testValue"

	// Set the value
	err := cache.Set(context.Background(), key, value)
	assert.NoError(t, err)

	// Delete the value
	assert.NotPanics(t, func() {
		cache.Del(context.Background(), key)
	})

	// Attempt to retrieve the deleted value
	var retrievedValue string
	found, err := cache.Get(context.Background(), key, &retrievedValue)
	assert.NoError(t, err)
	assert.False(t, found)
}

func TestInMemCacheAdapter_Get_UnregisteredCollection(t *testing.T) {
	cache := NewInMem()

	key := CollectionKey{
		Name: "unregisteredCollection",
		Key:  "testKey",
	}

	var retrievedValue string

	// Attempting to get from an unregistered collection should panic
	assert.Panics(t, func() {
		_, _ = cache.Get(context.Background(), key, &retrievedValue)
	})
}

func TestInMemCacheAdapter_Set_InvalidType(t *testing.T) {
	cache := NewInMem()

	coll := Collection{
		Name:          "testCollection",
		ValueType:     CollectionValueTypeString,
		LocalMaxItems: 10,
		Ttl:           0,
	}
	cache.Register(context.Background(), coll)

	key := CollectionKey{
		Name: "testCollection",
		Key:  "testKey",
	}

	value := 12345 // Invalid type for CollectionValueTypeString

	// Setting an invalid value type should panic
	assert.Panics(t, func() {
		_ = cache.Set(context.Background(), key, value)
	})
}

func TestInMemCacheAdapter_LocalMaxItems(t *testing.T) {
	cache := NewInMem()

	coll := Collection{
		Name:          "testCollection",
		ValueType:     CollectionValueTypeInt,
		LocalMaxItems: 5, // Set a limit of 5 items
		Ttl:           0,
	}
	cache.Register(context.Background(), coll)

	// Add 6 items to the cache, exceeding the LocalMaxItems limit
	keys := make([]CollectionKey, 6)
	for i := int64(0); i < 6; i++ {
		keys[i] = CollectionKey{
			Name: "testCollection",
			Key:  fmt.Sprintf("testKey-%d", i),
		}
		err := cache.Set(context.Background(), keys[i], i)
		assert.NoError(t, err)
	}

	// Validate the number of retained items
	retainedKeys := map[string]int64{}
	for _, key := range keys {
		var retrievedValue int64
		found, err := cache.Get(context.Background(), key, &retrievedValue)
		assert.NoError(t, err)
		if found {
			retainedKeys[key.Key] = retrievedValue
		}
	}

	// Ensure the number of retained keys does not exceed LocalMaxItems
	assert.LessOrEqual(t, len(retainedKeys), 5, "The number of retained items should not exceed LocalMaxItems")

	// Validate that the retained values are consistent
	for _, value := range retainedKeys {
		assert.Contains(t, []int64{0, 1, 2, 3, 4, 5}, value, "Retained value should be one of the set values")
	}
}
