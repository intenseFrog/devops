package common

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
)

func Confirm(msg string) bool {
	fmt.Println(msg)
	fmt.Print("Are you sure? (y/n): ")

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()

	if scanner.Text() == "y" {
		return true
	}

	return false
}

func Destroy(names []string, yes bool) {
	msg := fmt.Sprintf("About to remove %s", strings.Join(names, ", "))
	if !yes && !Confirm(msg) {
		return
	}

	for _, name := range names {
		node := &Node{Name: name}
		if err := node.Destroy(); err != nil {
			fmt.Println(err.Error())
		}
	}
}

func elite(args ...string) string {
	fmt.Printf("%s %s\n", config.Elite, strings.Join(args, " "))
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
