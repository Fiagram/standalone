package configs

import "time"

type CORS struct {
	IsEnable         bool          `yaml:"isEnable"`
	AllowOrigins     []string      `yaml:"allowOrigins"`
	AllowMethods     []string      `yaml:"allowMethods"`
	AllowHeaders     []string      `yaml:"allowHeaders"`
	ExposeHeaders    []string      `yaml:"exposeHeaders"`
	AllowCredentials bool          `yaml:"allowCredentials"`
	MaxAge           time.Duration `yaml:"maxAge"`
}

type Http struct {
	Address string `yaml:"address"`
	Port    string `yaml:"port"`
	CORS    CORS   `yaml:"CORS"`
}

func GetConfigHttp(c Config) Http {
	return c.Http
}
