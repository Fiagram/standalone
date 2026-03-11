package configs

type DatabaseClientType string

const DatabaseTypeMySql DatabaseClientType = "mysql"

type DatabaseClient struct {
	Type     DatabaseClientType `yaml:"type"`
	Address  string             `yaml:"address"`
	Port     int                `yaml:"port"`
	Username string             `yaml:"username"`
	Password string             `yaml:"password"`
	Database string             `yaml:"database"`
}

func GetConfigDatabaseClient(c Config) DatabaseClient {
	return c.DatabaseClient
}
