package katredis

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/mobiletoly/gokatana/katapp"
	"github.com/mobiletoly/gokatana/katcache"
	"github.com/redis/go-redis/v9"
	"io"
	"log"
	"log/slog"
)

var _ katcache.Cache = (*RedisCache)(nil)

type RedisCache struct {
	client *redis.Client
	chunks map[string]redisCacheChunk
	logger *slog.Logger
}

type redisCacheChunk struct {
	coll katcache.Collection
}

func NewRedisCache(ctx context.Context, addr string, DB int) (*RedisCache, error) {
	logger := katapp.Logger(ctx).WithGroup("RedisCache")
	logger.InfoContext(ctx, fmt.Sprintf("Open redis client for cache addr=%s db=%d", addr, DB))
	client := redis.NewClient(&redis.Options{
		Addr: addr,
		DB:   DB,
	})
	status := client.Ping(context.Background())
	if status.Err() != nil {
		logger.ErrorContext(ctx, "failed to connect to redis", "error", status.Err())
		return nil, fmt.Errorf("failed to connect to redis: %w", status.Err())
	}
	return &RedisCache{
		client: client,
		chunks: make(map[string]redisCacheChunk),
		logger: logger.Logger,
	}, nil
}

func (adp *RedisCache) Close() {
	if adp.client != nil {
		_ = adp.client.Close()
	}
}

// Register must be called before any other methods, such as Set, Get, Del.
func (adp *RedisCache) Register(ctx context.Context, coll katcache.Collection) {
	adp.logger.InfoContext(ctx, "Register cache collection", "collection", coll.Name)
	cn := coll.Name
	if _, ok := adp.chunks[cn]; ok {
		log.Panicf("Cache collection with name=%s was already registered", cn)
	}
	adp.chunks[cn] = redisCacheChunk{
		coll: coll,
	}
}

func (adp *RedisCache) mustGetCacheChunk(ns string) redisCacheChunk {
	if chunk, ok := adp.chunks[ns]; ok {
		return chunk
	}
	log.Panicf("Cache collection with name=%s was not registered", ns)
	return redisCacheChunk{}
}

func (adp *RedisCache) Set(ctx context.Context, ck katcache.CollectionKey, value any) error {
	chunk := adp.mustGetCacheChunk(ck.Name)

	var data []byte
	switch chunk.coll.ValueType {
	case katcache.CollectionValueTypeObject:
		data = chunk.coll.ValueType.MustSerialize(value)
	case katcache.CollectionValueTypeInt:
		data = []byte(fmt.Sprintf("%d", value.(int64)))
	case katcache.CollectionValueTypeReal:
		data = []byte(fmt.Sprintf("%f", value.(float64)))
	case katcache.CollectionValueTypeString:
		data = []byte(value.(string))
	default:
		log.Panicf("unknown value type: %d", chunk.coll.ValueType)
	}

	ctx = context.WithoutCancel(ctx)
	rk := buildRedisCacheKey(ck)
	cmd := adp.client.Set(ctx, rk, data, chunk.coll.Ttl)
	if err := cmd.Err(); err != nil {
		adp.logger.ErrorContext(ctx, "set record to redis failed", "key", rk, "cause", err)
		return fmt.Errorf("set record to redis by key=%s failed with error: %v", rk, err)
	} else {
		adp.logger.DebugContext(ctx, "successfully set data to redis", "key", rk)
	}
	return nil
}

func (adp *RedisCache) Get(ctx context.Context, ck katcache.CollectionKey, value any) (bool, error) {
	chunk := adp.mustGetCacheChunk(ck.Name)
	rk := buildRedisCacheKey(ck)
	cmd := adp.client.Get(ctx, rk)
	valueData, err := cmd.Bytes()
	if err != nil {
		if !errors.Is(err, redis.Nil) {
			adp.logger.ErrorContext(ctx, "read record from redis failed", "key", rk, "cause", err)
			return false, fmt.Errorf("read redis data by key=%s failed with error: %v", rk, err)
		}
		return false, nil
	}

	switch chunk.coll.ValueType {
	case katcache.CollectionValueTypeObject:
		chunk.coll.ValueType.MustDeserialize(valueData, value)
	case katcache.CollectionValueTypeInt:
		// parse convert valueData as string and parse it as int64
		_, err := fmt.Sscanf(string(valueData), "%d", value)
		if err != nil {
			panic(fmt.Errorf("failed to parse int64 from string: %v", err))
		}
	case katcache.CollectionValueTypeReal:
		// parse convert valueData as string and parse it as float64
		_, err := fmt.Sscanf(string(valueData), "%f", value)
		if err != nil {
			panic(fmt.Errorf("failed to parse float64 from string: %v", err))
		}
	case katcache.CollectionValueTypeString:
		*value.(*string) = string(valueData)
	default:
		log.Panicf("unknown value type: %d", chunk.coll.ValueType)
	}

	return true, nil
}

func (adp *RedisCache) Del(ctx context.Context, ck katcache.CollectionKey) error {
	rk := buildRedisCacheKey(ck)
	cmd := adp.client.Del(ctx, rk)
	if err := cmd.Err(); err != nil {
		adp.logger.ErrorContext(ctx, "delete item in redis failed", "key", rk, "cause", err)
		return fmt.Errorf("delete item in redis by key=%s failed with error: %w", rk, err)
	} else {
		return nil
	}
}

// buildRedisCacheKey creates a key that contains two parts - namespace and key itself.
// If namespace+key is shorter than 100 characters then it will be stored in a basic format such as "namespace:key",
// otherwise the key part will be encoded in SHA1. This is due to the fact that long keys under some
// circumstances can reduce cache performance (e.g. this is the case with Redis).
func buildRedisCacheKey(ck katcache.CollectionKey) string {
	var newKey string
	if len(ck.Name)+len(ck.Key) < 100 {
		newKey = ck.Key
	} else {
		h := sha1.New()
		_, _ = io.WriteString(h, ck.Key)
		newKey = hex.EncodeToString(h.Sum(nil))
	}
	return fmt.Sprintf("%s:%s", ck.Name, newKey)
}
