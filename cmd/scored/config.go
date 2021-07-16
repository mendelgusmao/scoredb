package main

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

var ScoreDB Specification

type Specification struct {
	Listen  string `default:":7363"`
	Logging bool   `default:"false"`
}

func readConfig() error {
	if err := envconfig.Process("scoredb", &ScoreDB); err != nil {
		return fmt.Errorf("readConfig: %s", err)
	}

	return nil
}
