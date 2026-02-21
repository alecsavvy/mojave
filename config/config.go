package config

// Config holds application configuration (e.g. from app.toml).
type Config struct{}

// DefaultConfig returns the default configuration.
func DefaultConfig() Config {
	return Config{}
}

// Load returns configuration from the given path. For now returns default and nil error.
func Load(path string) (*Config, error) {
	cfg := DefaultConfig()
	return &cfg, nil
}
