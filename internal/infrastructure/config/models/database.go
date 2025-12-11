package models

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Type        string `mapstructure:"DB_TYPE" default:"sqlite"`
	Echo        bool   `mapstructure:"DB_ECHO" default:"false"`
	Timeout     int    `mapstructure:"DB_TIMEOUT" default:"60"`
	WALEnable   bool   `mapstructure:"DB_WAL_ENABLE" default:"true"`
	PoolType    string `mapstructure:"DB_POOL_TYPE" default:"QueuePool"`
	PoolPrePing bool   `mapstructure:"DB_POOL_PRE_PING" default:"true"`
	PoolRecycle int    `mapstructure:"DB_POOL_RECYCLE" default:"300"`
	PoolTimeout int    `mapstructure:"DB_POOL_TIMEOUT" default:"30"`

	// SQLite 配置
	SQLitePoolSize    int `mapstructure:"DB_SQLITE_POOL_SIZE" default:"10"`
	SQLiteMaxOverflow int `mapstructure:"DB_SQLITE_MAX_OVERFLOW" default:"50"`

	// PostgreSQL 配置
	PostgreSQLHost        string `mapstructure:"DB_POSTGRESQL_HOST" default:"localhost"`
	PostgreSQLPort        int    `mapstructure:"DB_POSTGRESQL_PORT" default:"5432"`
	PostgreSQLDatabase    string `mapstructure:"DB_POSTGRESQL_DATABASE" default:"moviepilot"`
	PostgreSQLUsername    string `mapstructure:"DB_POSTGRESQL_USERNAME" default:"moviepilot"`
	PostgreSQLPassword    string `mapstructure:"DB_POSTGRESQL_PASSWORD" default:"moviepilot"`
	PostgreSQLPoolSize    int    `mapstructure:"DB_POSTGRESQL_POOL_SIZE" default:"10"`
	PostgreSQLMaxOverflow int    `mapstructure:"DB_POSTGRESQL_MAX_OVERFLOW" default:"50"`
}
