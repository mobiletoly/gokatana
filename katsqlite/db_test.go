package katsqlite

import (
	"github.com/mobiletoly/gokatana/kattest"
	"os"
	"testing"

	"github.com/mobiletoly/gokatana/katapp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConnect_Success(t *testing.T) {
	ctx := kattest.AppTestContext()

	// Create a temporary file for the SQLite database
	tempFile, err := os.CreateTemp("", "test_db_*.sqlite")
	require.NoError(t, err)
	defer os.Remove(tempFile.Name())

	cfg := &katapp.DatabaseConfig{
		File: tempFile.Name(),
	}

	// Connect to the database
	dbLink, err := Connect(ctx, cfg)
	require.NoError(t, err)
	assert.NotNil(t, dbLink)

	// Ensure the database connection works
	err = dbLink.Ping()
	assert.NoError(t, err)

	// Clean up
	dbLink.Close()
}

func TestConnect_InvalidPath(t *testing.T) {
	ctx := kattest.AppTestContext()

	cfg := &katapp.DatabaseConfig{
		File: "/invalid/path/test_db.sqlite",
	}

	// Attempt to connect to the database
	dbLink, err := Connect(ctx, cfg)
	assert.Error(t, err)
	assert.Nil(t, dbLink)
}

func TestClose(t *testing.T) {
	ctx := kattest.AppTestContext()

	// Create a temporary file for the SQLite database
	tempFile, err := os.CreateTemp("", "test_db_*.sqlite")
	require.NoError(t, err)
	defer os.Remove(tempFile.Name())

	cfg := &katapp.DatabaseConfig{
		File: tempFile.Name(),
	}

	// Connect to the database
	dbLink, err := Connect(ctx, cfg)
	require.NoError(t, err)
	assert.NotNil(t, dbLink)

	// Close the database
	dbLink.Close()

	// Attempt to close again and ensure no error is raised
	dbLink.Close()
}
