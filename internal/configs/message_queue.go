package configs

type MessageQueue struct {
	Addresses []string `yaml:"addresses"`
	ClientID  string   `yaml:"clientID"`
}

func GetConfigMessageQueue(c Config) MessageQueue {
	return c.MessageQueue
}
