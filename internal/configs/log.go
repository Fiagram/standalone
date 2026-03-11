package configs

type Log struct {
	Level string `yaml:"level"`
}

func GetConfigLog(c Config) Log {
	return c.Log
}
