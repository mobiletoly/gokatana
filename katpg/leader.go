package katpg

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// LeaderElector provides methods to manage leader election using Postgres advisory locks and pgx.
type LeaderElector struct {
	pool            *pgxpool.Pool
	conn            *pgxpool.Conn // pinned session
	lockKey         int64
	checkPeriod     time.Duration
	heartbeatPeriod time.Duration

	cancel   context.CancelFunc
	isLeader bool
	errCh    chan error
}

// NewLeaderElector creates a LeaderElector for a specific advisory lock key.
// - pool: a *pgxpool.Pool connection pool
// - lockKey: the Postgres advisory lock key (int64)
// - checkPeriod: how often to acquire lock if not leader
// - heartbeatPeriod: how often to confirm we still hold the lock if leader
func NewLeaderElector(pool *pgxpool.Pool, lockKey int64, checkPeriod, heartbeatPeriod time.Duration) *LeaderElector {
	return &LeaderElector{
		pool:            pool,
		lockKey:         lockKey,
		checkPeriod:     checkPeriod,
		heartbeatPeriod: heartbeatPeriod,
		errCh:           make(chan error, 1),
	}
}

// Start begins competing for leadership in a background goroutine.
// It returns a channel for receiving errors (e.g., DB errors).
func (le *LeaderElector) Start(ctx context.Context) <-chan error {
	ctx, le.cancel = context.WithCancel(ctx)

	// Acquire exactly one connection from the pool
	conn, err := le.pool.Acquire(ctx)
	if err != nil {
		// If we can’t get a connection, we can’t proceed.
		le.errCh <- fmt.Errorf("failed to acquire connection for leader election: %w", err)
		return le.errCh
	}
	le.conn = conn

	go le.run(ctx)
	return le.errCh
}

// Stop ends the leader election process and releases the lock if held.
func (le *LeaderElector) Stop() {
	if le.cancel != nil {
		le.cancel()
	}
}

// IsLeader returns whether we currently believe we are leader.
func (le *LeaderElector) IsLeader() bool {
	return le.isLeader
}

// run is the background routine that handles acquiring and confirming leadership.
func (le *LeaderElector) run(ctx context.Context) {
	acquireTicker := time.NewTicker(le.checkPeriod)
	heartbeatTicker := time.NewTicker(le.heartbeatPeriod)
	defer func() {
		acquireTicker.Stop()
		heartbeatTicker.Stop()
		if le.conn != nil {
			le.conn.Release()
		}
		close(le.errCh)
	}()

	for {
		select {
		case <-ctx.Done():
			if le.isLeader {
				_ = le.unlock()
			}
			return

		// Periodically try to acquire if not leader
		case <-acquireTicker.C:
			if !le.isLeader {
				acquired, err := le.tryLock(ctx)
				if err != nil {
					le.errCh <- fmt.Errorf("failed to acquire advisory lock: %w", err)
				} else if acquired {
					le.isLeader = true
				}
			}

		// Periodically confirm we still hold the lock if we are leader
		case <-heartbeatTicker.C:
			if le.isLeader {
				stillHeld, err := le.checkLockHeld(ctx)
				if err != nil {
					// If the check fails, we give up leadership
					//le.errCh <- fmt.Errorf("heartbeat check failed: %w", err)
					_, _ = fmt.Fprintf(os.Stderr, "1. heartbeat check failed: %v", err)
					le.isLeader = false
				} else if !stillHeld {
					_, _ = fmt.Fprintf(os.Stderr, "2. heartbeat check failed: %v", err)
					// We lost the lock unexpectedly
					le.isLeader = false
				}
			}
		}
	}
}

// tryLock attempts to acquire the advisory lock (non-blocking) using pg_try_advisory_lock.
func (le *LeaderElector) tryLock(ctx context.Context) (bool, error) {
	var acquired bool
	row := le.conn.QueryRow(ctx, "SELECT pg_try_advisory_lock($1)", le.lockKey)
	if err := row.Scan(&acquired); err != nil {
		return false, err
	}
	return acquired, nil
}

// unlock runs the unlock query with a fresh context (because original context may be canceled by this point).
func (le *LeaderElector) unlock() error {
	ctx := context.Background()
	_, err := le.conn.Exec(ctx, "SELECT pg_advisory_unlock($1)", le.lockKey)
	if err != nil {
		return fmt.Errorf("failed to release advisory lock: %w", err)
	}
	return err
}

func (le *LeaderElector) checkLockHeld(ctx context.Context) (bool, error) {
	var count int
	err := le.conn.QueryRow(ctx,
		`
        SELECT COUNT(*)
        FROM pg_locks
        WHERE locktype = 'advisory'
          AND objid = $1
          AND pid = pg_backend_pid();
        `,
		le.lockKey,
	).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
