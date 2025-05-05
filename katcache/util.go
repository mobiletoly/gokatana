package katcache

import (
	"context"
	"fmt"
	"github.com/mobiletoly/gokatana/katapp"
)

func LoadWithConfig(ctx context.Context, cfg *katapp.CacheConfig) (Cache, func()) {
	switch cfg.Type {
	case "none":
		return NewNone(), func() {}
	case "inmem":
		return NewInMem(), func() {}
	//case "redis":
	//	r := waveredis.NewRedis(ctx, cfg.Redis.Addr, cfg.Redis.DB)
	//	return r, func() {
	//		r.Close()
	//	}
	default:
		panic(fmt.Sprintf("unknown cache type: %s", cfg.Type))
	}
}

//func Get[T any](ctx context.Context, cache Cache, ck CollectionKey) *T {
//	value := new(T)
//	if cache.Get(ctx, ck, &value) {
//		return value
//	}
//	return nil
//}
//
//func Set[T any](ctx context.Context, cache Cache, ck CollectionKey, value *T) {
//	cache.Set(ctx, ck, value)
//}

// GetOrBuild fetches value from cache first and returns it if cache has it already
// Otherwise builder callback function will be invoked
//func GetOrBuild[T any](
//	ctx context.Context, cache Cache, ck CollectionKey,
//	builder func() (value *T, doNotSave bool, err error),
//) (*T, error) {
//	value := new(T)
//	if cache.Get(ctx, ck, value) {
//		return value, nil
//	}
//	newValue, doNotSave, err := builder()
//	if err != nil {
//		return nil, err
//	}
//	if !doNotSave {
//		cache.Set(ctx, ck, newValue)
//	}
//	return newValue, nil
//}
