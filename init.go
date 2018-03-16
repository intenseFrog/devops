package main

import (
	"log"
	"os"
	"path/filepath"
)

var config *Config

type Config struct {
	DirBaseImages string

	Create  string
	License string

	SSHPass string

	DirQcow2 string
}

func NewConfig(baseDIR string) *Config {
	config := &Config{
		DirBaseImages: baseDIR + "/base_images",
		DirQcow2:      baseDIR + "/qcow2",
		Create:        baseDIR + "/create_vms_2d.sh",
		License:       baseDIR + "/chiwen-license",
		SSHPass:       "sshpass -p sihua!@#890",
	}

	return config
}

func init() {
	myEnvDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}

	config = NewConfig(myEnvDir)
}
