package common

import (
	"io/ioutil"
	"sort"
	"strings"

	"gopkg.in/yaml.v2"
)

type Deployment struct {
	Myctl struct {
		Image   string `yaml:"image"`
		Channel string `yaml:"channel"`
	} `yaml:"myctl"`
	Nodes []*Node `yaml:"nodes"`

	master *Node
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

func ParseDeployment(path string) (*Deployment, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return parseDeployment(content)
}

func parseDeployment(data []byte) (*Deployment, error) {
	d := &Deployment{}
	if err := yaml.Unmarshal(data, d); err != nil {
		return nil, err
	}

	for _, n := range d.Nodes {
		if n.Role == "master" {
			d.master = n
			break
		}
	}

	sort.Slice(d.Nodes, func(i, j int) bool {
		iNode, jNode := d.Nodes[i], d.Nodes[j]
		if iNode.Role == "master" {
			return true
		} else if jNode.Role == "master" {
			return false
		}

		iValue := iNode.Cluster + iNode.Role
		jValue := jNode.Cluster + jNode.Role

		// Use < 0 to make sort in place
		return strings.Compare(iValue, jValue) < 0
	})

	return d, nil
}
