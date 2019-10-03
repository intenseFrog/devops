package pkg

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

const (
	RoleMaster  = "master"
	RoleLeader  = "leader"
	RoleManager = "manager"
	RoleWorker  = "worker"
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
		h := &Host{Name: name}
		if err := h.Delete(); err != nil {
			log.Debug(err.Error())
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

func output(r io.Reader, output chan<- string, done chan<- bool) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		output <- scanner.Text()
	}
	done <- true
}

func Exec(name string, args ...string) *exec.Cmd {
	cmds := append([]string{name}, args...)
	log.Debug(strings.Join(cmds, " "))
	return exec.Command(name, args...)
}

// CmdOutput: gives stdout, stderr, error
func Output(cmd *exec.Cmd) (string, string) {
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

	defer stderr.Close()
	defer stdout.Close()

	stdoutCh := make(chan string)
	stderrCh := make(chan string)
	doneCh := make(chan bool)

	go output(stdout, stdoutCh, doneCh)
	go output(stderr, stderrCh, doneCh)

	stdoutBuilder := strings.Builder{}
	stderrBuilder := strings.Builder{}

	for i := 2; i > 0; {
		select {
		case s := <-stdoutCh:
			fmt.Println(s)
			if _, err := stdoutBuilder.WriteString(s + "\n"); err != nil {
				panic(err)
			}
		case e := <-stderrCh:
			fmt.Println(e)
			if _, err := stderrBuilder.WriteString(e + "\n"); err != nil {
				panic(err)
			}
		case <-doneCh:
			i--
		}
	}

	return stdoutBuilder.String(), stderrBuilder.String()
}
