package katsqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/mobiletoly/gokatana/katapp"
	"log"
	"log/slog"
	"time"

	"github.com/mobiletoly/gokatana/katcache"
)

var _ katcache.Cache = (*KVTCache)(nil)

type KVTCache struct {
	db              *sql.DB
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

// NewKVTCache creates a new SQLite-based KVTCache.
func NewKVTCache(
	ctx context.Context,
	dbFile string,
	approveDeletion func(ctx context.Context, coll katcache.Collection) bool,
) (*KVTCache, error) {
	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		return nil, fmt.Errorf("failed to open SQLite database: %w", err)
	}
	// Optimize connection settings for a web server environment
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	_, err = db.Exec("PRAGMA journal_mode=WAL;")
	if err != nil {
		return nil, fmt.Errorf("failed to set journal mode to WAL: %w", err)
	}

	return &KVTCache{
		db:              db,
		chunks:          make(map[string]dbCacheChunk),
		logger:          katapp.Logger(ctx).WithGroup("KVTCache").Logger,
		approveDeletion: approveDeletion,
	}, nil
}

func (c *KVTCache) Close() {
	_ = c.db.Close()
	c.logger.Info("cache closed")
}

// mustGetCacheChunk retrieves the cache chunk for a specific collection.
// Panics if the chunk is not registered.
func (c *KVTCache) mustGetCacheChunk(ctx context.Context, ns string) dbCacheChunk {
	if chunk, ok := c.chunks[ns]; ok {
		return chunk
	}
	c.logger.ErrorContext(ctx, "collection not registered", "namespace", ns)
	log.Panicf("cache chunk for namespace %s was not registered", ns)
	return dbCacheChunk{}
}

// Register creates a cache table for the specified collection.
// It will panic if the table creation fails or collection was already registered.
// It is important to call this method to register all collections before
// start using the cache (and prior to Run methods).
func (c *KVTCache) Register(ctx context.Context, coll katcache.Collection) {
	table := coll.Name
	ttl := int64(coll.Ttl.Seconds())

	createTableSQL := fmt.Sprintf(`
CREATE TABLE IF NOT EXISTS %s (
	key TEXT PRIMARY KEY,
	value BLOB NOT NULL,
	added_at DATETIME DEFAULT CURRENT_TIMESTAMP
);`, table)

	_, err := c.db.ExecContext(ctx, createTableSQL)
	if err != nil {
		log.Panicf("failed to create cache table for collection %s: %v", table, err)
	}

	deleteExpiredSQL := fmt.Sprintf(`
DELETE FROM %s WHERE added_at < DATETIME('now', '-%d seconds');`, table, ttl)

	upsertValueSQL := fmt.Sprintf(`
INSERT INTO %s (key, value, added_at)
VALUES (?, ?, CURRENT_TIMESTAMP)
ON CONFLICT(key) DO UPDATE SET value=excluded.value, added_at=CURRENT_TIMESTAMP;`, table)

	selectByKeySQL := fmt.Sprintf(`
SELECT value FROM %s WHERE key = ? AND added_at >= DATETIME('now', '-%d seconds');`, table, ttl)

	deleteByKeySQL := fmt.Sprintf(`
DELETE FROM %s WHERE key = ?;`, table)

	c.chunks[table] = dbCacheChunk{
		coll:             coll,
		deleteExpiredSql: deleteExpiredSQL,
		upsertValueSql:   upsertValueSQL,
		selectByKeySql:   selectByKeySQL,
		deleteByKeySql:   deleteByKeySQL,
	}

	c.logger.InfoContext(ctx, "registered collection", "collection", coll.Name)
}

// Run starts periodic cleanup of expired rows for a specific collection.
func (c *KVTCache) Run(ctx context.Context, coll katcache.Collection) {
	c.logger.InfoContext(ctx, "starting cache runner", "collection", coll.Name)

	chunk := c.mustGetCacheChunk(ctx, coll.Name)
	ticker := time.NewTicker(coll.Ttl)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if c.approveDeletion == nil || c.approveDeletion(ctx, coll) {
				_, err := c.db.ExecContext(ctx, chunk.deleteExpiredSql)
				if err != nil {
					c.logger.ErrorContext(ctx, "failed to delete expired rows", "collection", coll.Name, "error", err)
				} else {
					c.logger.DebugContext(ctx, "deleted expired rows", "collection", coll.Name)
				}
			}
		case <-ctx.Done():
			c.logger.InfoContext(ctx, "stopping cache runner", "collection", coll.Name)
			return
		}
	}
}

// RunInBackgroundWithCancellation starts cache runners for all collections and returns a cancel function.
func (c *KVTCache) RunInBackgroundWithCancellation(ctx context.Context) (cancel func()) {
	ctx, cancelFunc := context.WithCancel(ctx)

	for _, chunk := range c.chunks {
		go func(coll katcache.Collection) {
			c.Run(ctx, coll)
		}(chunk.coll)
	}

	return cancelFunc
}

// Set stores a key-value pair in the cache.
func (c *KVTCache) Set(ctx context.Context, ck katcache.CollectionKey, value any) error {
	chunk := c.mustGetCacheChunk(ctx, ck.Name)
	serializedValue := chunk.coll.ValueType.MustSerialize(value)

	_, err := c.db.ExecContext(ctx, chunk.upsertValueSql, ck.Key, serializedValue)
	if err != nil {
		return fmt.Errorf("failed to set value in cache: %w", err)
	}
	return nil
}

// Get retrieves a value from the cache.
func (c *KVTCache) Get(ctx context.Context, ck katcache.CollectionKey, value any) (bool, error) {
	chunk := c.mustGetCacheChunk(ctx, ck.Name)

	row := c.db.QueryRowContext(ctx, chunk.selectByKeySql, ck.Key)
	var serializedValue []byte
	err := row.Scan(&serializedValue)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to retrieve value: %w", err)
	}

	chunk.coll.ValueType.MustDeserialize(serializedValue, value)
	return true, nil
}

// Del deletes a key from the cache.
func (c *KVTCache) Del(ctx context.Context, ck katcache.CollectionKey) error {
	chunk := c.mustGetCacheChunk(ctx, ck.Name)

	_, err := c.db.ExecContext(ctx, chunk.deleteByKeySql, ck.Key)
	if err != nil {
		return fmt.Errorf("failed to delete key from cache: %w", err)
	}
	return nil
}
