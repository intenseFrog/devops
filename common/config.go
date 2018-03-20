package common

import (
	"log"
	"os"
	"path/filepath"
)

var config *Config

type Config struct {
	DirBaseImages string

	Create  string
	Elite   string
	License string

	SSHPass string

	DirQcow2 string
}

func NewConfig(baseDIR string) *Config {
	config := &Config{
		DirBaseImages: baseDIR + "/base_images",
		DirQcow2:      baseDIR + "/qcow2",
		Create:        baseDIR + "/scripts/create_vms_2d.sh",
		Elite:         baseDIR + "/scripts/elite",
		License:       baseDIR + "/scripts/chiwen-license",
		SSHPass:       `sshpass -p 'sihua!@#890'`,
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
