package pkg

import (
	"strings"
	"time"
)

type MyArguments struct {
	commands []myArg
}

type myArg struct {
	output bool
	args   []string
}

func (e *MyArguments) Append(output bool, args ...string) {
	if len(e.commands) == 0 {
		e.commands = make([]myArg, 0)
	}

	e.commands = append(e.commands, myArg{output: output, args: args})
}

func my(args ...string) (string, string) {
	return Output(Exec(config.My, args...))
}

// FIXME: supposed to evaluate my stderr
func myLogin(masterIP string, timeout time.Duration) {
	until := time.Now().Add(timeout)
	for {
		if !time.Now().Before(until) {
			break
		}
		stdout, _ := my("login", "-u", "admin", "-p", "admin", masterIP)
		if !strings.Contains(stdout, "ERROR") {
			break
		}
		time.Sleep(5 * time.Second)
	}
}

func myLogout() {
	my("logout")
}
