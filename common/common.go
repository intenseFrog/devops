package common

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
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

func PrettyDuration(d time.Duration) string {
	var result string

	total := int(d.Seconds())
	units := []string{"s", "m"}
	base := 60

	for i := range units {
		if total < 60 {
			return fmt.Sprintf("%d%s%s", total, units[i], result)
		}

		remainder := total % base
		if remainder != 0 {
			result = fmt.Sprintf("%d%s%s", remainder, units[i], result)
		}

		total = total / base
	}

	if total != 0 {
		result = fmt.Sprintf("%dh%s", total, result)
	}

	return result
}

func PrintDone(start time.Time) {
	fmt.Printf("Done: %s\n", PrettyDuration(time.Now().Sub(start)))
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

func RemoveKnownHosts() {
	Output(exec.Command("rm", "~/.ssh/known_hosts"))
}
