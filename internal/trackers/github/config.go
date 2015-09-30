package github

import (
	// Stdlib
	"log"

	// Vendor
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	// Story label.
	StoryLabel string `envconfig:"GH_STORY_LABEL" default:"story"`

	// State labels.
	ApprovedLabel         string `envconfig:"GH_APPROVED_LABEL"          default:"approved"`
	BeingImplementedLabel string `envconfig:"GH_BEING_IMPLEMENTED_LABEL" default:"being implemented"`
	ImplementedLabel      string `envconfig:"GH_IMPLEMENTED_LABEL"       default:"implemented"`
	ReviewedLabel         string `envconfig:"GH_REVIEWED_LABEL"          default:"reviewed"`
	SkipReviewLabel       string `envconfig:"GH_SKIP_REVIEW_LABEL"       default:"no review"`
	PassedTestingLabel    string `envconfig:"GH_PASSED_TESTING_LABEL"    default:"qa+"`
	FailedTestingLabel    string `envconfig:"GH_FAILED_TESTING_LABEL"    default:"qa-"`
	SkipTestingLabel      string `envconfig:"GH_SKIP_TESTING_LABEL"      default:"no qa"`
	StagedLabel           string `envconfig:"GH_STAGED_LABEL"            default:"staged"`
	RejectedLabel         string `envconfig:"GH_REJECTED_LABEL"          default:"rejected"`
}

var config Config

func init() {
	if err := envconfig.Process("SFD", &config); err != nil {
		log.Fatalln("Fatal error while parsing GitHub config:", err)
	}
}

func GetConfig() Config {
	return config
}
