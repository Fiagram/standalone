package configs

import (
	"net/http"
	"strings"
	"time"
)

type Http struct {
	Address string `yaml:"address"`
	Port    string `yaml:"port"`
	CORS    CORS   `yaml:"CORS"`
	Cookie  Cookie `yaml:"cookie"`
}

type CORS struct {
	IsEnable         bool          `yaml:"isEnable"`
	AllowOrigins     []string      `yaml:"allowOrigins"`
	AllowMethods     []string      `yaml:"allowMethods"`
	AllowHeaders     []string      `yaml:"allowHeaders"`
	ExposeHeaders    []string      `yaml:"exposeHeaders"`
	AllowCredentials bool          `yaml:"allowCredentials"`
	MaxAge           time.Duration `yaml:"maxAge"`
}

type Cookie struct {
	Domain       string `yaml:"domain"`
	SameSiteMode string `yaml:"sameSite"`
	Path         string `yaml:"path"`
	Secure       bool   `yaml:"secure"`
	HttpOnly     bool   `yaml:"httpOnly"`
}

func (c Cookie) SameSite() http.SameSite {
	switch strings.ToLower(c.SameSiteMode) {
	case "strict":
		return http.SameSiteStrictMode
	case "lax":
		return http.SameSiteLaxMode
	case "none":
		return http.SameSiteNoneMode
	default:
		return http.SameSiteDefaultMode
	}
}

func GetConfigHttp(c Config) Http {
	return c.Http
}

func GetConfigHttpCORS(c Config) CORS {
	return c.Http.CORS
}

func GetConfigHttpCookie(c Config) Cookie {
	return c.Http.Cookie
}
