package main

import (
	"bufio"
	"os"
)

type Deployment struct {
	Nodes []*Node
	Myctl string

	Master *Node
}

func (d *Deployment) Create() error {
	for _, n := range d.Nodes {
		if err := n.Create(); err != nil {
			return err
		}
	}

	return nil
}

func (d *Deployment) Destroy() error {
	for _, n := range d.Nodes {
		if err := n.Destroy(); err != nil {
			return err
		}
	}

	return nil
}

// TODO: does YAML make more sense?
func Parse(path string) (*Deployment, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Scan()

	deployment := &Deployment{Myctl: scanner.Text(), Nodes: make([]*Node, 0)}

	for scanner.Scan() {
		node := &Node{Deployment: deployment}
		if err := node.Parse(scanner.Text()); err != nil {
			return nil, err
		}

		if node.Role == "master" {
			deployment.Master = node
		}

		deployment.Nodes = append(deployment.Nodes, node)
	}

	return deployment, nil
}
