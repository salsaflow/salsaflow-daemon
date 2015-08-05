package pivotaltracker

import (
	// Stdlib
	"log"
	"os"

	// Vendor
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Token               string `envconfig:"PT_TOKEN"`
	ReviewedLabel       string `envconfig:"PT_REVIEWED_LABEL"`
	ReviewSkippedLabel  string `envconfig:"PT_REVIEW_SKIPPED_LABEL"`
	TestingPassedLabel  string `envconfig:"PT_TESTING_PASSED_LABEL"`
	TestingFailedLabel  string `envconfig:"PT_TESTING_FAILED_LABEL"`
	TestingSkippedLabel string `envconfig:"PT_TESTING_SKIPPED_LABEL"`
}

var config Config

func init() {
	if err := envconfig.Process("SFD", &config); err != nil {
		log.Fatalln("Fatal error while parsing Pivotal Tracker config:", err)
	}

	var missing bool
	ensure := func(variable, value string) {
		if value == "" {
			log.Printf("environment variable not set: %v\n", variable)
			missing = true
		}
	}

	ensure("SFD_PT_TOKEN", config.Token)
	ensure("SFD_PT_REVIEWED_LABEL", config.ReviewedLabel)
	ensure("SFD_PT_REVIEW_SKIPPED_LABEL", config.ReviewSkippedLabel)
	ensure("SFD_PT_TESTING_PASSED_LABEL", config.TestingPassedLabel)
	ensure("SFD_PT_TESTING_FAILED_LABEL", config.TestingFailedLabel)
	ensure("SFD_PT_TESTING_SKIPPED_LABEL", config.TestingSkippedLabel)

	if missing {
		os.Exit(1)
	}
}

func GetConfig() Config {
	return config
}
