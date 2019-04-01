package common

import (
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type eliteArg struct {
	output bool
	args   []string
}

type EliteArguments struct {
	commands []eliteArg
}

func (e *EliteArguments) Append(output bool, args ...string) {
	if len(e.commands) == 0 {
		e.commands = make([]eliteArg, 0)
	}

	e.commands = append(e.commands, eliteArg{output: output, args: args})
}

func elite(args ...string) (string, string) {
	fmt.Printf("%s %s\n", config.Elite, strings.Join(args, " "))
	return Output(exec.Command(config.Elite, args...))
}

// FIXME: supposed to evaluate elite stderr
func eliteLogin(masterIP string, timeout time.Duration) {
	until := time.Now().Add(timeout)
	for {
		if !time.Now().Before(until) {
			break
		}
		stdout, _ := elite("login", "-u", "admin", "-p", "admin", masterIP)
		if !strings.Contains(stdout, "ERROR") {
			break
		}
		time.Sleep(5 * time.Second)
	}
}

func eliteLogout() {
	elite("logout")
}
