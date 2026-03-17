package configs

import "time"

type Auth struct {
	Hash  Hash  `yaml:"hash"`
	Token Token `yaml:"token"`
}

type Hash struct {
	Cost int `yaml:"cost"`
}

type Token struct {
	Secret              string        `yaml:"secret"`
	AccessTokenTTL      time.Duration `yaml:"accessTokenTTL"`
	RefreshTokenLongTTL time.Duration `yaml:"refreshTokenLongTTL"`
	RefreshTokenTTL     time.Duration `yaml:"refreshTokenTTL"`
}

func GetConfigAuth(c Config) Auth {
	return c.Auth
}

func GetConfigAuthToken(c Config) Token {
	return c.Auth.Token
}

func GetConfigAuthHash(c Config) Hash {
	return c.Auth.Hash
}
