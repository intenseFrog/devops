package common

import (
	"log"
	"os"
	"path/filepath"
)

var config *Config

type Config struct {
	Elite   string
	License string
}

func NewConfig(baseDIR string) *Config {
	config := &Config{
		Elite:   baseDIR + "/scripts/elite",
		License: baseDIR + "/scripts/chiwen-license",
	}

	return config
}

func init() {
	baseDIR, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}

	config = NewConfig(baseDIR)
}
