package katpg

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mobiletoly/gokatana/katapp"
	"github.com/mobiletoly/gokatana/katcache"
	"log"
	"log/slog"
	"time"
)

var _ katcache.Cache = (*KVTCache)(nil)

type KVTCache struct {
	db              *pgxpool.Pool
	schema          string
	chunks          map[string]dbCacheChunk
	logger          *slog.Logger
	approveDeletion func(ctx context.Context, coll katcache.Collection) bool
}

type dbCacheChunk struct {
	coll             katcache.Collection
	deleteExpiredSql string
	upsertValueSql   string
	selectByKeySql   string
	deleteByKeySql   string
}

type cacheRecord struct {
	ID      int       `db:"id"`
	Key     string    `db:"key"`
	Value   []byte    `db:"value"`
	AddedAt time.Time `db:"added_at"`
}

func (c *KVTCache) Close() {
	c.db.Close()
}

// NewKVTCache creates a new KVTCache instance.
//   - schema: the database schema for the cache tables to be created in
//   - approveDeletion: a function to approve deletion of expired data. If function is nil or if true is returned,
//     all expired data will be deleted. If false is returned, no data will be deleted. This can be useful
//     in combination with a leader election mechanism to ensure that only one instance of the application
//     deletes expired data.
func NewKVTCache(
	ctx context.Context,
	db *pgxpool.Pool,
	schema string,
	approveDeletion func(ctx context.Context, coll katcache.Collection) bool,
) (*KVTCache, error) {
	kvt := &KVTCache{
		db:              db,
		schema:          schema,
		chunks:          make(map[string]dbCacheChunk),
		logger:          katapp.Logger(ctx).WithGroup("KVTCache").Logger,
		approveDeletion: approveDeletion,
	}
	return kvt, nil
}

// Register creates a new cache table for a specific collection.
// It will panic if the table creation fails or collection was already registered.
// It is important to call this method to register all collections before
// start using the cache (and prior to Run methods).
func (c *KVTCache) Register(ctx context.Context, coll katcache.Collection) {
	c.logger.InfoContext(ctx, "registering cache for collection", "collection", coll)

	createKVTTableSqlFormat := `
CREATE UNLOGGED TABLE IF NOT EXISTS %s.%s (
	id BIGSERIAL PRIMARY KEY,
	key TEXT NOT NULL UNIQUE,
	value BYTEA NOT NULL,
	added_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS %s_%s__added_at ON %s.%s (added_at);
`
	c.logger.Debug("create table", "collection", coll.Name, "sql", createKVTTableSqlFormat)

	table := coll.Name
	createSql := fmt.Sprintf(createKVTTableSqlFormat, c.schema, table, c.schema, table, c.schema, table)
	err := pgx.BeginFunc(ctx, c.db, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, createSql)
		return err
	})
	if err != nil {
		log.Panicf("failed to create KVT table: %v", err)
	}

	expiration := int64(coll.Ttl.Seconds())
	deleteExpiredSql := fmt.Sprintf(`
DELETE FROM %s.%s
WHERE added_at < NOW() - INTERVAL '%d seconds'
`, c.schema, table, expiration)

	upsertValueSql := fmt.Sprintf(`
INSERT INTO %s.%s(key, value, added_at)
VALUES ($1, $2, NOW())
ON CONFLICT (key) DO UPDATE
    SET value    = EXCLUDED.value,
        added_at = NOW()
`, c.schema, table)

	selectByKeySql := fmt.Sprintf(`
SELECT id, key, value, added_at
FROM %s.%s
WHERE key = $1 AND added_at > NOW() - INTERVAL '%d seconds'
LIMIT 1
`, c.schema, table, expiration)

	deleteByKeySql := fmt.Sprintf(`
DELETE FROM %s.%s
WHERE key = $1
`, c.schema, table)

	c.chunks[table] = dbCacheChunk{
		coll:             coll,
		deleteExpiredSql: deleteExpiredSql,
		upsertValueSql:   upsertValueSql,
		selectByKeySql:   selectByKeySql,
		deleteByKeySql:   deleteByKeySql,
	}
}

// Run starts a blocking cache runner for a specific collection.
// It will delete expired data from the cache based on TTL.
// Cancelling the context will stop the runner.
func (c *KVTCache) Run(
	ctx context.Context,
	coll katcache.Collection,
) {
	c.logger.InfoContext(ctx, "starting cache for collection", "collection", coll.Name)
	chunk := c.mustGetCacheChunk(ctx, coll.Name)
	_, err := c.db.Exec(ctx, chunk.deleteExpiredSql)
	if err != nil {
		log.Panicf("failed to delete expired data from collection=%s: %v", coll.Name, err)
	}
	ticker := time.NewTicker(coll.Ttl)
	defer ticker.Stop()

loop:
	for {
		select {
		case <-ticker.C:
			if c.approveDeletion == nil || c.approveDeletion(ctx, coll) {
				c.logger.DebugContext(ctx, "deleting expired data",
					"collection", coll.Name, "ttl", coll.Ttl.String())
				_, err := c.db.Exec(ctx, chunk.deleteExpiredSql)
				if err != nil {
					c.logger.ErrorContext(ctx, "failed to delete expired data from cache",
						"collection", coll.Name, "cause", err)
				}
			}
		case <-ctx.Done():
			c.logger.DebugContext(ctx, "cache collection context was cancelled", "collection", coll.Name)
			break loop
		}
	}
	c.logger.InfoContext(ctx, "stopped cache for collection", "collection", coll.Name)
}

// RunInBackgroundWithCancellation starts a cache runner for each registered collection in the background.
// It returns a cancel function to stop the runners.
func (c *KVTCache) RunInBackgroundWithCancellation(ctx context.Context) (cancel func()) {
	ctx, cancel = context.WithCancel(ctx)
	for _, chunk := range c.chunks {
		go func() {
			c.Run(ctx, chunk.coll)
		}()
	}
	return cancel
}

// Set stores a value in the cache.
func (c *KVTCache) Set(ctx context.Context, ck katcache.CollectionKey, value any) error {
	chunk := c.mustGetCacheChunk(ctx, ck.Name)
	data := chunk.coll.ValueType.MustSerialize(value)
	ctx = context.WithoutCancel(ctx)
	_, err := c.db.Exec(ctx, chunk.upsertValueSql, ck.Key, data)
	if err != nil {
		return fmt.Errorf("failed to set data to cache by key=%s (ignored): %w", ck.String(), err)
	}
	return nil
}

// Get retrieves a value from the cache.
// Returns true if the key was found in the cache, false otherwise.
// If the key was found, the value is deserialized into the provided value pointer.
func (c *KVTCache) Get(ctx context.Context, ck katcache.CollectionKey, value any) (bool, error) {
	chunk := c.mustGetCacheChunk(ctx, ck.Name)
	rows, _ := c.db.Query(ctx, chunk.selectByKeySql, ck.Key)
	rec, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[cacheRecord])
	if IsNoRows(err) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to get data from cache by key=%s (ignored): %w", ck.String(), err)
	}
	chunk.coll.ValueType.MustDeserialize(rec.Value, value)
	return true, nil
}

// Del deletes a value from the cache.
func (c *KVTCache) Del(ctx context.Context, ck katcache.CollectionKey) error {
	chunk := c.mustGetCacheChunk(ctx, ck.Name)
	ctx = context.WithoutCancel(ctx)
	_, err := c.db.Exec(ctx, chunk.deleteByKeySql, ck.Key)
	if err != nil {
		return fmt.Errorf("failed to delete data from cache by key=%s (ignored): %w", ck.String(), err)
	}
	return nil
}

func (c *KVTCache) mustGetCacheChunk(ctx context.Context, ns string) dbCacheChunk {
	if chunk, ok := c.chunks[ns]; ok {
		return chunk
	}
	log.Panicf("CacheConfig partition with namespace=%s was not registered", ns)
	return dbCacheChunk{}
}
