package configs

import (
	"fmt"
	"os"

	"github.com/Fiagram/standalone/configs"
	"gopkg.in/yaml.v3"
)

type ConfigFilePath string

type Config struct {
	Http           Http            `yaml:"http"`
	MessageQueue   MessageQueue    `yaml:"message_queue"`
	Auth           Auth            `yaml:"auth"`
	Log            Log             `yaml:"log"`
	CacheClient    CacheClient     `yaml:"cache_client"`
	DatabaseClient DatabaseClient  `yaml:"database_client"`
	GrpcClient     GrpcClient      `yaml:"grpc_client"`
	Strategy       StrategyFeature `yaml:"strategy"`
}

// Creates a new config instance by reading from a given YAML file.
// If the filePath is empty, it uses the default embedded configuration.
func NewConfig(filePath ConfigFilePath) (Config, error) {
	var (
		configBytes = configs.DefaultConfigBytes
		config      = Config{}
		err         error
	)

	if filePath != "" {
		configBytes, err = os.ReadFile(string(filePath))
		if err != nil {
			return Config{}, fmt.Errorf("Failed to read YAML file: %w", err)
		}
	}

	err = yaml.Unmarshal(configBytes, &config)
	if err != nil {
		return Config{}, fmt.Errorf("Failed to unmarshal YAML file: %w", err)
	}

	return config, nil
}
