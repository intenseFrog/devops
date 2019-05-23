package common

import "sync"

type Node struct {
	Name string `yaml:"name"`
	Role string `yaml:"role"`

	cluster *Cluster
}

var joinMutex sync.Mutex

func (n *Node) Join(hostID string) {
	joinMutex.Lock()
	defer joinMutex.Unlock()
	my("cluster", "use", n.cluster.Name)
	my("cluster", "join", hostID, "--role", n.Role)
}
