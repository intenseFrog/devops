package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os/exec"
	"strings"
)

func Elite(args ...string) string {
	fmt.Printf("\n\n\ninit: %s %s\n\n\n", config.Elite, strings.Join(args, " "))
	stdout, stderr := Output(exec.Command(config.Elite, args...))
	if stderr != "" {
		fmt.Println(stderr)
	}

	return stdout

}

// CmdOutput: gives stdout, stderr, error
func Output(cmd *exec.Cmd) (outStr string, errStr string) {
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Fatal(err)
	}

	if err = cmd.Start(); err != nil {
		log.Fatal(err)
	}

	outPut, err := ioutil.ReadAll(stdout)
	if err != nil {
		log.Fatal(err)
	}
	outStr = strings.Trim(string(outPut), "\n")

	errOutput, err := ioutil.ReadAll(stderr)
	if err != nil {
		log.Fatal(err)
	}
	errStr = strings.Trim(string(errOutput), "\n")

	return
}
