package katpg

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMain sets up any global test requirements (e.g., logging).
func TestMain(m *testing.M) {
	// You can do additional setup here if needed.
	// Then run tests:
	code := m.Run()
	os.Exit(code)
}

func TestLeaderElection_WithPgxPool(t *testing.T) {
	// 1. Spin up a PostgreSQL container.
	bctx := context.Background()
	pc := RunPostgresTestContainer(bctx, t, nil, nil)
	t.Cleanup(func() {
		pc.Terminate(bctx, t)
	})

	// 4. Create a pgx pool.
	pool := pc.BuildPgxPool(bctx, t)
	defer pool.Close()

	// Create a LeaderElector.
	elector := NewLeaderElector(
		pool,
		1234,          // advisory lock key
		1*time.Second, // checkPeriod
		2*time.Second, // heartbeatPeriod
	)

	// Start the elector.
	ctx := context.WithValue(bctx, "elector", "elector-1")
	errCh := elector.Start(ctx)
	defer elector.Stop()

	// Listen for errors in a separate goroutine (optional).
	go func() {
		for e := range errCh {
			log.Printf("LeaderElector error: %v", e)
		}
	}()

	// Wait for the elector to (likely) become leader.
	var becameLeader bool
	for i := 0; i < 10; i++ {
		if elector.IsLeader() {
			becameLeader = true
			break
		}
		time.Sleep(300 * time.Millisecond)
	}
	require.True(t, becameLeader, "elector did not become leader within 3 seconds")

	// 9. Basic assertion that we are indeed the leader.
	assert.True(t, elector.IsLeader(), "we should hold leadership")

	// 10. Optionally, create a second elector to ensure the first remains leader.
	elector2 := NewLeaderElector(
		pool,
		1234,          // same advisory lock key
		1*time.Second, // checkPeriod
		2*time.Second, // heartbeatPeriod
	)
	ctx = context.WithValue(bctx, "elector", "elector-2")
	errCh2 := elector2.Start(ctx)
	defer elector2.Stop()

	// Wait a little for the second elector to attempt acquiring the lock
	time.Sleep(2 * time.Second)

	// The first elector should still be leader, second elector not leader
	assert.True(t, elector.IsLeader(), "first elector should remain leader")
	assert.False(t, elector2.IsLeader(), "second elector should not be leader")

	// Stop the first elector; second should become leader eventually
	elector.Stop()

	var secondBecameLeader bool
	for i := 0; i < 10; i++ {
		if elector2.IsLeader() {
			secondBecameLeader = true
			break
		}
		time.Sleep(300 * time.Millisecond)
	}
	assert.True(t, secondBecameLeader, "second elector did not become leader after first stopped")
	elector2.Stop()

	// Print any errors from second elector (if any)
	for e := range errCh2 {
		t.Logf("LeaderElector2 error: %v", e)
	}
}
