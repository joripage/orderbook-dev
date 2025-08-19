package config

import (
	"os"

	postgres_wrapper "github.com/joripage/orderbook-dev/pkg/infra/postgres"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

type AppConfig struct {
	ServiceName string                           `yaml:"service_name"`
	OmsDB       *postgres_wrapper.PostgresConfig `yaml:"oms_db"`
}

// Load load config from file and environment variables.
func Load(filePath string) (*AppConfig, error) {
	if len(filePath) == 0 {
		filePath = os.Getenv("CONFIG_FILE")
	}

	fields := []interface{}{
		"func",
		"config.readFromFile",
		"filePath",
		filePath,
	}

	sugar := zap.S().With(fields...)

	sugar.Debug("Load config...")
	zap.S().Debugf("CONFIG_FILE=%v", filePath)

	configBytes, err := os.ReadFile(filePath)
	if err != nil {
		sugar.Error("Failed to load config file")
		return nil, err
	}
	configBytes = []byte(os.ExpandEnv(string(configBytes)))

	cfg := &AppConfig{}

	err = yaml.Unmarshal(configBytes, cfg)
	if err != nil {
		sugar.Error("Failed to parse config file")
		return nil, err
	}

	zap.S().Debugf("config: %+v", cfg)

	return cfg, nil
}
