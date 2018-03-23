package common

import (
	"fmt"
	"sort"
)

type Cluster struct {
	Name   string            `yaml:"name"`
	Kind   string            `yaml:"kind"`
	Params map[string]string `yaml:"parameters,omitempty"`
	Nodes  []*Node           `yaml:"nodes"`

	deployment *Deployment
}

func (c *Cluster) CleanKnownHosts() error {
	for _, node := range c.Nodes {
		node.CleanKnownHost()
	}

	return nil
}

func (c *Cluster) Create() error {
	for _, node := range c.Nodes {
		if err := node.Create(); err != nil {
			return err
		}
	}

	return nil
}

func (c *Cluster) deployDefault() error {
	for _, node := range c.Nodes[1:] {
		if err := node.swarmNode().join(); err != nil {
			return err
		}
	}

	return nil
}

func (c *Cluster) deployKubernetes() error {
	leader := c.Nodes[0].kubernetesNode()
	if err := leader.init(); err != nil {
		return err
	}

	for _, node := range c.Nodes[1:] {
		if err := node.kubernetesNode().join(); err != nil {
			return err
		}
	}

	return nil
}

func (c *Cluster) deploySwarm() error {
	leader := c.Nodes[0].swarmNode()
	if err := leader.init(); err != nil {
		return err
	}

	for _, node := range c.Nodes[1:] {
		if err := node.swarmNode().join(); err != nil {
			return err
		}
	}

	return nil
}

func (c *Cluster) Deploy() (err error) {
	if c.Name == "default" {
		err = c.deployDefault()
	} else {
		switch c.Kind {
		case "swarm":
			err = c.deploySwarm()
		case "kubernetes":
			err = c.deployKubernetes()
		default:
			err = fmt.Errorf("invalid kind of cluster: %s", c.Kind)
		}
	}

	return
}

func (c *Cluster) Destroy() {
	for _, node := range c.Nodes {
		node.Destroy()
	}
}

func (c *Cluster) masterIP() string {
	return c.deployment.masterIP()
}

func (c *Cluster) myctlImage() string {
	return c.deployment.myctlImage()
}

func (c *Cluster) myctlChannel() string {
	return c.deployment.myctlChannel()
}

// Sort nodes in the order of role: master > leader > worker
// assign cluster to each node
// return master if found
func (c *Cluster) Normalize() (master *Node) {
	for i, node := range c.Nodes {
		if node.Role == "master" {
			master = c.Nodes[i]
		}
		node.cluster = c
	}

	sort.Slice(c.Nodes, func(i, j int) bool {
		iNode, jNode := c.Nodes[i], c.Nodes[j]
		if iNode.Role == "master" {
			return true
		} else if jNode.Role == "master" {
			return false
		}

		if iNode.Role == "leader" {
			return true
		} else if jNode.Role == "leader" {
			return false
		}

		// must both be workers
		return true
	})

	return
}
