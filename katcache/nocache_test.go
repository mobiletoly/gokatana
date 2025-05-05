package katcache

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNoCacheAdapter_Register(t *testing.T) {
	cache := NewNone()

	coll := Collection{
		Name:          "testNamespace",
		ValueType:     CollectionValueTypeObject,
		Ttl:           0,
		LocalMaxItems: 0,
	}

	// Register a partition
	assert.NotPanics(t, func() {
		cache.Register(context.Background(), coll)
	})

	// Attempt to register the same namespace again, should panic
	assert.Panics(t, func() {
		cache.Register(context.Background(), coll)
	})
}

func TestNoCacheAdapter_Set(t *testing.T) {
	cache := NewNone()
	ctx := context.Background()

	coll := Collection{
		Name:          "testNamespace",
		ValueType:     CollectionValueTypeObject,
		Ttl:           0,
		LocalMaxItems: 0,
	}
	cache.Register(context.Background(), coll)

	ck := CollectionKey{Name: "testNamespace", Key: "testKey"}

	// Set a value in the cache (does nothing in NoCacheAdapter)
	assert.NotPanics(t, func() {
		_ = cache.Set(ctx, ck, "testValue")
	})

	// Verify the key is not retrievable
	found, err := cache.Get(ctx, ck, nil)
	assert.NoError(t, err)
	assert.False(t, found)
}

func TestNoCacheAdapter_Get(t *testing.T) {
	cache := NewNone()
	ctx := context.Background()

	coll := Collection{
		Name:          "testNamespace",
		ValueType:     CollectionValueTypeObject,
		Ttl:           0,
		LocalMaxItems: 0,
	}
	cache.Register(context.Background(), coll)

	key := CollectionKey{Name: "testNamespace", Key: "testKey"}

	// Attempt to get a value (always returns false in NoCacheAdapter)
	found, err := cache.Get(ctx, key, nil)
	assert.NoError(t, err)
	assert.False(t, found)
}

func TestNoCacheAdapter_Del(t *testing.T) {
	cache := NewNone()
	ctx := context.Background()

	coll := Collection{
		Name:          "testNamespace",
		ValueType:     CollectionValueTypeObject,
		Ttl:           0,
		LocalMaxItems: 0,
	}
	cache.Register(context.Background(), coll)

	key := CollectionKey{Name: "testNamespace", Key: "testKey"}

	// Delete a key (does nothing in NoCacheAdapter)
	assert.NotPanics(t, func() {
		cache.Del(ctx, key)
	})
}

func TestNoCacheAdapter_RegisterAndMustHaveCacheChunk(t *testing.T) {
	cache := NewNone()
	adapter := cache.(*NoCache)

	coll := Collection{
		Name:          "testNamespace",
		Ttl:           0,
		LocalMaxItems: 0,
	}
	cache.Register(context.Background(), coll)
	assert.NotPanics(t, func() {
		adapter.mustHaveCacheChunk("testNamespace")
	})
	assert.Panics(t, func() {
		adapter.mustHaveCacheChunk("unregisteredNamespace")
	})
}

func TestNoCacheAdapter_Close(t *testing.T) {
	cache := NewNone()

	// Close the cache (does nothing in NoCacheAdapter)
	assert.NotPanics(t, func() {
		cache.Close()
	})
}
