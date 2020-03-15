package config

import (
	"os"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

// Config contains the parsed contents of hover.yaml
type Config struct {
	loaded   bool
	Relution struct {
		Host  string
		Token string
	}
	Mysql struct {
		Host     string
		Port     int
		User     string
		Password string
		Name     string
	}
	Port int
}

var config = Config{}

// GetConfig returns the working directory config.yaml as a Config
func GetConfig() *Config {
	if !config.loaded {
		c, err := ReadConfigFile("config.yaml")
		if err != nil {
			return &config
		}
		config = *c
		config.loaded = true
	}
	return &config
}

// ReadConfigFile reads a .yaml file at a path and return a correspond
// Config struct
func ReadConfigFile(configPath string) (*Config, error) {
	file, err := os.Open(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.Wrap(err, "Warning: No hover.yaml file found")
		}
		return nil, errors.Wrap(err, "Failed to open hover.yaml")
	}
	defer file.Close()

	var config Config
	err = yaml.NewDecoder(file).Decode(&config)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to decode hover.yaml")
	}
	return &config, nil
}
