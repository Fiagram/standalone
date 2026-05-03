package configs

import "time"

type SePayConfig struct {
	APIKey        string        `yaml:"api_key"`
	AccountNumber string        `yaml:"account_number"`
	BankCode      string        `yaml:"bank_code"`
	WebhookSecret string        `yaml:"webhook_secret"`
	PaymentExpiry time.Duration `yaml:"payment_expiry"`
}

func GetConfigSePay(c Config) SePayConfig {
	return c.SePay
}
