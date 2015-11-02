package github

import (
	// Stdlib
	"log"

	// Vendor
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	WebhookSecret string `envconfig:"WEBHOOK_SECRET"`
	Token         string `envconfig:"TOKEN"`
}

var config Config

func init() {
	if err := envconfig.Process("SFD_GITHUB", &config); err != nil {
		log.Fatalln("Fatal error while parsing GitHub config:", err)
	}
}

func GetConfig() Config {
	return config
}
