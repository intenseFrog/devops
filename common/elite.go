package common

import (
	"fmt"
	"os/exec"
	"strings"
	"sync"
)

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

func elite(e *EliteArguments) (output []string) {
	eliteMutex.Lock()
	defer eliteMutex.Unlock()

	for _, cmd := range e.commands {
		fmt.Printf("%s %s\n", config.Elite, strings.Join(cmd.args, " "))
		stdout, stderr := Output(exec.Command(config.Elite, cmd.args...))
		if stderr != "" {
			fmt.Println(stderr)
		}

		if cmd.output {
			output = append(output, stdout)
		}
	}

	return output
}

func eliteLogin(masterIP string) {
	stdout, _ := Output(exec.Command(config.Elite, "login", "-u", "admin", "-p", "admin", masterIP))
	fmt.Println(stdout)
}

func eliteLogout() {
	stdout, _ := Output(exec.Command(config.Elite, "logout"))
	fmt.Println(stdout)
}
