package configs

type GrpcClient struct {
	Strategy Strategy `yaml:"strategy"`
}

type Strategy struct {
	Address string `yaml:"address"`
	Port    int    `yaml:"port"`
}

func GetConfigGrpcClient(c Config) GrpcClient {
	return c.GrpcClient
}

func GetConfigGrpcClientStrategy(c Config) Strategy {
	return c.GrpcClient.Strategy
}
