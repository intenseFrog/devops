package common

import (
	"fmt"
	"os/exec"
	"strings"
	"sync"
)

// global mutex
var eliteMutex sync.Mutex

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

func elite(args ...string) string {
	eliteMutex.Lock()
	defer eliteMutex.Unlock()

	fmt.Printf("%s %s\n", config.Elite, strings.Join(args, " "))
	stdout, _ := Output(exec.Command(config.Elite, args...))
	return stdout
}

func eliteLogin(masterIP string) {
	Output(exec.Command(config.Elite, "login", "-u", "admin", "-p", "admin", masterIP))
}

func eliteLogout() {
	Output(exec.Command(config.Elite, "logout"))
}
