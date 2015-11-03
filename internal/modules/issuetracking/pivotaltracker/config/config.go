package config

import (
	// Stdlib
	"log"

	// Vendor
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	WebhookSecret string `envconfig:"WEBHOOK_SECRET"`
	Token         string `envconfig:"TOKEN"`

	ReviewedLabel       string `envconfig:"REVIEWED_LABEL"        default:"reviewed"`
	ReviewSkippedLabel  string `envconfig:"EVIEW_SKIPPED_LABEL"  default:"no review"`
	TestingPassedLabel  string `envconfig:"TESTING_PASSED_LABEL"  default:"qa+"`
	TestingFailedLabel  string `envconfig:"TESTING_FAILED_LABEL"  default:"qa-"`
	TestingSkippedLabel string `envconfig:"TESTING_SKIPPED_LABEL" default:"no qa"`
}

var config Config

func init() {
	if err := envconfig.Process("SFD_PIVOTALTRACKER", &config); err != nil {
		log.Fatalln("Fatal error while parsing Pivotal Tracker config:", err)
	}
}

func Get() Config {
	return config
}
