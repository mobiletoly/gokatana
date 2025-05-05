package katapp_test

import (
	"github.com/mobiletoly/gokatana/katapp"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func createTempYamlFile(t *testing.T, content string) string {
	t.Helper()

	file, err := os.CreateTemp("", "*.yaml")
	if err != nil {
		t.Fatalf("failed to create temporary file: %v", err)
	}
	defer file.Close()

	_, err = file.WriteString(content)
	if err != nil {
		t.Fatalf("failed to write content to temporary file: %v", err)
	}

	return file.Name()
}

func TestLoad_CommonAndDeploymentConfig(t *testing.T) {
	commonContent := `
server:
  addr: "0.0.0.0"
  port: 8080
`
	deploymentContent := `
database:
  migrations:
    - service: "example-service"
      schema: "public"
      path: "./migrations"
  user: "admin"
  password: "secret"
`

	tmpDir := t.TempDir()
	commonFile := filepath.Join(tmpDir, "common.yaml")
	deploymentFile := filepath.Join(tmpDir, "deployment.yaml")

	// Write the content to temporary files
	_ = os.WriteFile(commonFile, []byte(commonContent), 0644)
	_ = os.WriteFile(deploymentFile, []byte(deploymentContent), 0644)

	deployment := katapp.Deployment{
		Name:            "deployment",
		ConfigDir:       tmpDir,
		CommonConfigDir: tmpDir,
	}

	type Config struct {
		Server   katapp.ServerConfig   `mapstructure:"server"`
		Database katapp.DatabaseConfig `mapstructure:"database"`
	}

	cfg := katapp.LoadConfig[Config]("", deployment)

	// Validate server configuration
	assert.Equal(t, "0.0.0.0", cfg.Server.Addr)
	assert.Equal(t, 8080, cfg.Server.Port)

	// Validate database configuration
	assert.Len(t, cfg.Database.Migrations, 1)
	assert.Equal(t, "example-service", cfg.Database.Migrations[0].Service)
	assert.Equal(t, "public", cfg.Database.Migrations[0].Schema)
	assert.Equal(t, "./migrations", cfg.Database.Migrations[0].Path)
	assert.Equal(t, "admin", cfg.Database.User)
	assert.Equal(t, "secret", cfg.Database.Password)
}

func TestLoad_MissingCommonConfig(t *testing.T) {
	deploymentContent := `
server:
  addr: "127.0.0.1"
  port: 9090
`

	tmpDir := t.TempDir()
	deploymentFile := filepath.Join(tmpDir, "deployment.yaml")
	os.WriteFile(deploymentFile, []byte(deploymentContent), 0644)

	deployment := katapp.Deployment{
		Name:            "deployment",
		ConfigDir:       tmpDir,
		CommonConfigDir: "", // Simulating missing common config directory
	}

	type Config struct {
		Server katapp.ServerConfig `mapstructure:"server"`
	}

	cfg := katapp.LoadConfig[Config]("", deployment)

	// Validate server configuration
	assert.Equal(t, "127.0.0.1", cfg.Server.Addr)
	assert.Equal(t, 9090, cfg.Server.Port)
}

func TestLoad_WithEnvironmentVariables(t *testing.T) {
	commonContent := `
server:
  addr: "${SERVER_ADDR}"
  port: ${SERVER_PORT}
`
	os.Setenv("SERVER_ADDR", "192.168.1.1")
	os.Setenv("SERVER_PORT", "8081")
	defer os.Unsetenv("SERVER_ADDR")
	defer os.Unsetenv("SERVER_PORT")

	tmpDir := t.TempDir()
	commonFile := filepath.Join(tmpDir, "common.yaml")
	os.WriteFile(commonFile, []byte(commonContent), 0644)

	deployment := katapp.Deployment{
		Name:            "common",
		ConfigDir:       tmpDir,
		CommonConfigDir: tmpDir,
	}

	type Config struct {
		Server katapp.ServerConfig `mapstructure:"server"`
	}

	cfg := katapp.LoadConfig[Config]("", deployment)

	// Validate server configuration
	assert.Equal(t, "192.168.1.1", cfg.Server.Addr)
	assert.Equal(t, 8081, cfg.Server.Port)
}

func TestLoad_InvalidConfigFile(t *testing.T) {
	invalidContent := `
server:
  addr: "0.0.0.0"
  port: "invalid"  # Port should be an integer
`

	tmpDir := t.TempDir()
	commonFile := filepath.Join(tmpDir, "common.yaml")
	os.WriteFile(commonFile, []byte(invalidContent), 0644)

	deployment := katapp.Deployment{
		Name:            "common",
		ConfigDir:       tmpDir,
		CommonConfigDir: tmpDir,
	}

	type Config struct {
		Server katapp.ServerConfig `mapstructure:"server"`
	}

	assert.Panics(t, func() {
		katapp.LoadConfig[Config]("", deployment)
	})
}

func TestLoad_MissingConfigDir(t *testing.T) {
	deployment := katapp.Deployment{
		Name:            "deployment",
		ConfigDir:       "./nonexistent-dir",
		CommonConfigDir: "./nonexistent-dir",
	}

	type Config struct {
		Server katapp.ServerConfig `mapstructure:"server"`
	}

	assert.Panics(t, func() {
		katapp.LoadConfig[Config]("", deployment)
	})
}
