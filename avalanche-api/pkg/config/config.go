package config

type Config struct {
	Port             int      `mapstructure:"port"`
	DatabaseURI      string   `mapstructure:"database_uri"`
	ExtractKeys      []string `mapstructure:"extract_keys"`
	MaxWorkers       int      `mapstructure:"max_workers"`
	MaxRequestSize   int      `mapstructure:"max_request_size"`
	BatchSize        int      `mapstructure:"batch_size"`
	DatabasePoolSize int      `mapstructure:"database_pool_size"`
}

const (
	DefaultPort             = 8080
	DefaultMaxWorkers       = 5
	DefaultMaxRequestSize   = 1000
	DefaultBatchSize        = 100
	DefaultDatabasePoolSize = 30
)
