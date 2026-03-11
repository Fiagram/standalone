package configs

type Http struct {
	Address string `yaml:"address"`
	Port    string `yaml:"port"`
}

func GetConfigHttp(c Config) Http {
	return c.Http
}
