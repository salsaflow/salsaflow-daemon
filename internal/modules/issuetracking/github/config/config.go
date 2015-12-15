package config

import (
	// Stdlib
	"log"
	"strings"

	// Vendor
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	// Story label.
	StoryLabelList string `envconfig:"STORY_LABELS" default:"enhancement,bug"`

	// StoryLabels contains parsed StoryLabelList.
	StoryLabels []string

	// State labels.
	ApprovedLabel         string `envconfig:"APPROVED_LABEL"          default:"approved"`
	BeingImplementedLabel string `envconfig:"BEING_IMPLEMENTED_LABEL" default:"being implemented"`
	ImplementedLabel      string `envconfig:"IMPLEMENTED_LABEL"       default:"implemented"`
	ReviewedLabel         string `envconfig:"REVIEWED_LABEL"          default:"reviewed"`
	SkipReviewLabel       string `envconfig:"SKIP_REVIEW_LABEL"       default:"no review"`
	PassedTestingLabel    string `envconfig:"PASSED_TESTING_LABEL"    default:"qa+"`
	FailedTestingLabel    string `envconfig:"FAILED_TESTING_LABEL"    default:"qa-"`
	SkipTestingLabel      string `envconfig:"SKIP_TESTING_LABEL"      default:"no qa"`
	StagedLabel           string `envconfig:"STAGED_LABEL"            default:"staged"`
	RejectedLabel         string `envconfig:"REJECTED_LABEL"          default:"rejected"`
}

var config Config

func init() {
	if err := envconfig.Process("SFD_GITHUB", &config); err != nil {
		log.Fatalln("Fatal error while parsing GitHub config:", err)
	}

	mp := func(xs []string, mapFunc func(string) string) []string {
		ss := make([]string, len(xs))
		for i, x := range xs {
			ss[i] = mapFunc(x)
		}
		return ss
	}

	config.StoryLabels = mp(strings.Split(config.StoryLabelList, ","), strings.TrimSpace)
}

func Get() Config {
	return config
}
