package katcache

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/mobiletoly/gokatana/katapp"
	"math"
	"time"
)

type CollectionValueType int

// Partition data type (based on iota)
const (
	// CollectionValueTypeObject works with any data types that can be serialized to JSON
	CollectionValueTypeObject CollectionValueType = iota
	// CollectionValueTypeInt works with int64 data types only
	CollectionValueTypeInt
	// CollectionValueTypeReal works with float64 data types only
	CollectionValueTypeReal
	// CollectionValueTypeString works with string data types only
	CollectionValueTypeString
)

// CollectionKey consists of two parts - collection name (can be your entity type) and key name itself
// Refer to BuildCacheKey() function is this file for more information.
type CollectionKey struct {
	Name string
	Key  string
}

func (ck CollectionKey) String() string {
	return fmt.Sprintf("{Name=%s Key=%s}", ck.Name, ck.Key)
}

// Cache declares generic cache interface to be implemented by real cache implementation (such as Redis for example)
type Cache interface {
	Close()
	Register(ctx context.Context, coll Collection)
	Set(ctx context.Context, ck CollectionKey, value any) error
	Get(ctx context.Context, ck CollectionKey, value any) (bool, error)
	Del(ctx context.Context, ck CollectionKey) error
}

type Collection struct {
	Name      string              // Namespace key prefix, must be unique amongst other namespaces
	Ttl       time.Duration       // Time-to-live, 0 means no expiration
	ValueType CollectionValueType // Data type of the value (object, int, real, string)

	// Max number of items in cache (in-memory cache only, for redis etc will be ignored)
	// 	it can be useful, it cache provides both persistent and quick local cache
	LocalMaxItems int
}

func (coll Collection) String() string {
	return fmt.Sprintf("{name=%s ttl=%s valueType=%d localMaxItems=%d}",
		coll.Name, coll.Ttl.String(), coll.ValueType, coll.LocalMaxItems)
}

func (cvt CollectionValueType) MustSerialize(value any) []byte {
	switch cvt {
	case CollectionValueTypeObject:
		if result, err := json.Marshal(value); err != nil {
			panic(fmt.Errorf("error marshalling value: %w", err))
		} else {
			return result
		}
	case CollectionValueTypeInt:
		b := make([]byte, 8)
		binary.LittleEndian.PutUint64(b, uint64(value.(int64))) // Use PutUint64 for better performance
		return b
	case CollectionValueTypeReal:
		b := make([]byte, 8)
		binary.LittleEndian.PutUint64(b, math.Float64bits(value.(float64))) // Use PutUint64 for better performance
		return b
	case CollectionValueTypeString:
		return []byte(value.(string))
	default:
		panic(fmt.Sprintf("unknown value type: %d", cvt))
	}
}

func (cvt CollectionValueType) MustDeserialize(data []byte, value any) {
	switch cvt {
	case CollectionValueTypeObject:
		if err := json.Unmarshal(data, value); err != nil {
			panic(fmt.Errorf("error unmarshalling value: %w", err))
		}
	case CollectionValueTypeInt:
		*value.(*int64) = int64(binary.LittleEndian.Uint64(data))
	case CollectionValueTypeReal:
		*value.(*float64) = math.Float64frombits(binary.LittleEndian.Uint64(data))
	case CollectionValueTypeString:
		*value.(*string) = string(data)
	default:
		panic(fmt.Sprintf("unknown value type: %d", cvt))
	}
}

func SetAsync[T any](ctx context.Context, c Cache, ck CollectionKey, value T) {
	ctx = context.WithoutCancel(ctx)
	go func() {
		err := c.Set(ctx, ck, value)
		if err != nil {
			katapp.Logger(ctx).Warn("failed to set data to cache", "key", ck.String(),
				"cause", err)
		}
	}()
}

func DelAsync(ctx context.Context, c Cache, ck CollectionKey) {
	ctx = context.WithoutCancel(ctx)
	go func() {
		err := c.Del(ctx, ck)
		if err != nil {
			katapp.Logger(ctx).Warn("failed to delete data from cache", "key", ck.String(), "cause", err)
		}
	}()
}

// GetOrSetAsync is a helper method that gets a value from the cache by key and if it's not found, it calls the setter
// function to get the value and sets it in the cache. It ignores database errors as much as it can (by the
// end of the day, it's a cache, so losing data is not that important).
func GetOrSetAsync[T any](ctx context.Context, c Cache, key CollectionKey, fetch func(key CollectionKey) (T, error)) (value T, err error) {
	found, err := c.Get(ctx, key, &value)
	if err != nil {
		return value, fmt.Errorf("failed to get data from cache by key=%s (ignored): %w", key, err)
	}
	if found {
		return value, nil
	}
	fetchedValue, err := fetch(key)
	if err != nil {
		return value, err
	}
	SetAsync(ctx, c, key, fetchedValue)
	return fetchedValue, nil
}
