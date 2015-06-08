package pivotaltracker

import (
	// Stdlib
	"log"
	"os"

	// Vendor
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Token              string `envconfig:"PT_TOKEN"`
	ReviewedLabel      string `envconfig:"PT_REVIEWED_LABEL"`
	TestingPassedLabel string `envconfig:"PT_TESTING_PASSED_LABEL"`
	TestingFailedLabel string `envconfig:"PT_TESTING_FAILED_LABEL"`
	ImplementedLabel   string `envconfig:"PT_IMPLEMENTED_LABEL"`
}

var config Config

func init() {
	if err := envconfig.Process("SFD", &config); err != nil {
		log.Fatalln("Fatal error while parsing Pivotal Tracker config:", err)
	}

	var missing bool
	if config.Token == "" {
		log.Println("variable not set: SFD_PT_TOKEN")
		missing = true
	}
	if config.ReviewedLabel == "" {
		log.Println("variable not set: SFD_PT_REVIEWED_LABEL")
		missing = true
	}
	if config.TestingPassedLabel == "" {
		log.Println("variable not set: SFD_PT_TESTING_PASSED_LABEL")
		missing = true
	}
	if config.TestingFailedLabel == "" {
		log.Println("variable not set: SFD_PT_TESTING_FAILED_LABEL")
		missing = true
	}
	if config.ImplementedLabel == "" {
		log.Println("variable not set: SFD_PT_IMPLEMENTED_LABEL")
		missing = true
	}

	if missing {
		os.Exit(1)
	}
}

func GetConfig() Config {
	return config
}
