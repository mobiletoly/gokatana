package katapp

type ServerConfig struct {
	// Addr is a network address to listen on
	Addr string
	// Port is a network port to listen on
	Port int
	// RequestDecompression is a type of decompression to be used on incoming requests (e.g. "gzip")
	RequestDecompression string
	// ResponseCompression is a type of compression to be used on outgoing responses (e.g. "gzip")
	ResponseCompression string
}

type DatabaseConfig struct {
	// File is a path to the database file (e.g. SQLite)
	File string
	// Host is a network address of the database server
	Host string
	// Port is a network port of the database server
	Port int
	// Name is a name of the database
	Name string
	// User is a name of the database user
	User string
	// Password is a password of the database user
	Password string
	// Sslmode is a mode of SSL connection to the database (e.g. "disable")
	Sslmode string
	// ConnectTimeout is a timeout for establishing a connection to the database
	ConnectTimeout int
	// Migrations is a list of database migrations to be performed when database is connected
	Migrations []DatabaseMigrationConfig
	// Parameters provides additional connection parameters
	Parameters map[string]string
}

// DatabaseMigrationConfig represents a database migration configuration
type DatabaseMigrationConfig struct {
	// Service is a name of the service that uses this migration
	Service string
	// Schema is a name of the database schema where the migration table is stored
	Schema string
	// Path is a path to the directory with migration files
	Path string
}

type CacheConfig struct {
	Type string
}
