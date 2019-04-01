package common

type Node struct {
	Name string `yaml:"name"`
	Role string `yaml:"role"`

	cluster *Cluster
}

func (n *Node) Join(hostID string) {
	elite("cluster", "use", n.cluster.Name)
	elite("cluster", "join", hostID, "--role", n.Role)
}
