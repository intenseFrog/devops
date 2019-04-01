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
	elite("cluster", "use", n.cluster.Name)
	elite("cluster", "join", hostID, "--role", n.Role)
}
