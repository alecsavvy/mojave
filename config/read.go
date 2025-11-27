package config

func ReadConfig(homeDir string) (*Config, error) {
	config := DefaultConfig()

	return config, nil
}
