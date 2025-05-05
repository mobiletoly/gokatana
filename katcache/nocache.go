package katcache

import (
	"context"
	"fmt"
)

var _ Cache = (*NoCache)(nil)

type NoCache struct {
	chunks map[string]struct{}
}

func NewNone() Cache {
	return &NoCache{
		chunks: make(map[string]struct{}),
	}
}

func (adp *NoCache) Close() {
	// Nothing to do
}

func (adp *NoCache) Set(_ context.Context, ck CollectionKey, _ any) error {
	adp.mustHaveCacheChunk(ck.Name)
	return nil
}

func (adp *NoCache) Get(_ context.Context, ck CollectionKey, _ any) (bool, error) {
	adp.mustHaveCacheChunk(ck.Name)
	return false, nil
}

func (adp *NoCache) Del(_ context.Context, ck CollectionKey) error {
	adp.mustHaveCacheChunk(ck.Name)
	return nil
}

func (adp *NoCache) Register(_ context.Context, coll Collection) {
	ns := coll.Name
	if _, ok := adp.chunks[ns]; ok {
		panic(fmt.Sprintf("cache collection with name=%s was already registered", ns))
	}
	adp.chunks[ns] = struct{}{}
}

func (adp *NoCache) mustHaveCacheChunk(ns string) {
	if _, ok := adp.chunks[ns]; !ok {
		panic(fmt.Sprintf("cache collection with name=%s was not registered", ns))
	}
}
