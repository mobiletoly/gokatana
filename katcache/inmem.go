package katcache

import (
	"context"
	"fmt"
	"github.com/mobiletoly/gokatana/katcache/internal"
	"time"
)

var _ Cache = (*InMemCache)(nil)

type InMemCache struct {
	chunks map[string]*inMemCacheChunk
}

type inMemCacheChunk struct {
	coll  Collection
	cache *internal.TinyLFU
}

func NewInMem() *InMemCache {
	return &InMemCache{
		chunks: make(map[string]*inMemCacheChunk),
	}
}

func (adp *InMemCache) Close() {
	// Nothing to do
}

func (adp *InMemCache) Register(_ context.Context, coll Collection) {
	collName := coll.Name
	if _, ok := adp.chunks[collName]; ok {
		panic(fmt.Sprintf("Cache partition with namespace=%s was already registered", collName))
	}
	var ttl time.Duration
	if coll.Ttl > 0 {
		ttl = coll.Ttl
	} else {
		ttl = 100 * 365 * 24 * time.Hour // almost forever
	}
	c := internal.NewTinyLFU(coll.LocalMaxItems, ttl)
	adp.chunks[collName] = &inMemCacheChunk{
		coll:  coll,
		cache: c,
	}
}

func (adp *InMemCache) mustGetCacheChunk(collName string) *inMemCacheChunk {
	if chunk, ok := adp.chunks[collName]; ok {
		return chunk
	}
	panic(fmt.Sprintf("Cache collection with name=%s was not registered", collName))
}

func (adp *InMemCache) Set(_ context.Context, ck CollectionKey, value any) error {
	chunk := adp.mustGetCacheChunk(ck.Name)
	data := chunk.coll.ValueType.MustSerialize(value)
	chunk.cache.Set(ck.Key, data)
	return nil
}

func (adp *InMemCache) Get(_ context.Context, ck CollectionKey, value any) (bool, error) {
	chunk := adp.mustGetCacheChunk(ck.Name)
	data, ok := chunk.cache.Get(ck.Key)
	if !ok {
		return false, nil
	}
	chunk.coll.ValueType.MustDeserialize(data, value)
	return true, nil
}

func (adp *InMemCache) Del(_ context.Context, ck CollectionKey) error {
	chunk := adp.mustGetCacheChunk(ck.Name)
	chunk.cache.Del(ck.Key)
	return nil
}
