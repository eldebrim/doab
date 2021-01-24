package cmd

import (
	"os"
)

type config struct {
	packages []string
}

func newConfig(packages []string) *config {
	conf := config{packages: packages}
	return &conf
}
