package main

import "mydevops/cmd"

// CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o mydevops

func main() {
	cmd.Execute()
}
