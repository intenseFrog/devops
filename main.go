package main

import (
	log "github.com/sirupsen/logrus"

	"mydevops/cmd"
)

// CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o mydevops

func main() {
	log.SetLevel(log.DebugLevel)
	cmd.Execute()
}
