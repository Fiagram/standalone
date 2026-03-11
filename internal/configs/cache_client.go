package configs

type CacheClientType string

const (
	CacheTypeRam   CacheClientType = "ram"
	CacheTypeRedis CacheClientType = "redis"
)

type CacheClient struct {
	Type     CacheClientType `yaml:"type"`
	Address  string          `yaml:"address"`
	Port     string          `yaml:"port"`
	Username string          `yaml:"username"`
	Password string          `yaml:"password"`
}

func GetConfigCacheClient(c Config) CacheClient {
	return c.CacheClient
}
