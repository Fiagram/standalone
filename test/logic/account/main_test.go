package logic_account_test

import (
	"log"
	"os"
	"testing"

	"github.com/Fiagram/standalone/internal/configs"
)

var config *configs.Config

func TestMain(m *testing.M) {
	// Use the default config to test database connection
	cf, err := configs.NewConfig("")
	if err != nil {
		log.Fatal("failed to init config default")
	}

	config = &cf

	os.Exit(m.Run())
}
