package pkg

import (
	"fmt"
	"sort"
	"strings"
)

type Cluster struct {
	Name   string            `yaml:"name"`
	Kind   string            `yaml:"kind"`
	Params map[string]string `yaml:"parameters,omitempty"`
	Nodes  []*Node           `yaml:"nodes"`

	deployment *Deployment
}

func (c *Cluster) Deploy() {
	// create cluster first
	createArgs := []string{"cluster", "create", c.Name, "--" + c.Kind}
	for k, v := range c.Params {
		createArgs = append(createArgs, "-p", fmt.Sprintf("%s=%s", k, v))
	}
	my(createArgs...)

	stdout, _ := my("host", "ls")
	hostDict := parseHostOutput(stdout)
	for _, n := range c.Nodes {
		id, ok := hostDict[n.Name]
		if !ok {
			panic(fmt.Errorf("cannot find host %s", n.Name))
		}
		n.Join(id)
	}
}

// parse host out to the format of Name->ID
// "1    luke183   ready   5 Cores   9.5 GiB   172.16.88.183"
// luke183 -> 1

func parseHostOutput(output string) map[string]string {
	rows := strings.Split(output, "\n")
	res := make(map[string]string)
	for i := 1; i < len(rows); i++ {
		cols := strings.Split(rows[i], " ")
		id := cols[0]
		for _, row := range cols[1:] {
			if row != "" {
				// name -> id
				res[row] = id
				break
			}
		}
	}

	return res
}

// Sort nodes in the order of role: master > leader > worker
// assign cluster to each node
func (c *Cluster) Normalize() {
	for i := range c.Nodes {
		c.Nodes[i].cluster = c
	}

	sort.Slice(c.Nodes, func(i, j int) bool {
		return c.Nodes[i].Role == RoleManager || c.Nodes[i].Role == RoleLeader
	})
}
