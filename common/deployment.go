package common

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type Deployment struct {
	Myctl struct {
		Image   string `yaml:"image"`
		Channel string `yaml:"channel"`
	} `yaml:"myctl"`
	Clusters []*Cluster `yaml:"clusters"`

	master *Node
}

func (d *Deployment) CleanKnownHosts() {
	for _, c := range d.Clusters {
		c.CleanKnownHosts()
	}
}

func (d *Deployment) Create() error {
	for _, c := range d.Clusters {
		if err := c.Create(); err != nil {
			return err
		}
	}

	return nil
}

func (d *Deployment) Deploy() error {
	defer elite("logout")

	for i := range d.Clusters {
		cluster := d.Clusters[i]
		cluster.deployment = d

		if master := cluster.normalize(); master != nil {
			d.master = master
		}
	}

	fmt.Println("Licensing...")
	if err := d.master.License(); err != nil {
		return err
	}

	fmt.Println("Deploying master...")
	if err := d.master.Deploy(); err != nil {
		return err
	}

	fmt.Println("Deploying clusters...")
	for _, c := range d.Clusters {
		if err := c.Deploy(); err != nil {
			return err
		}
	}

	return nil
}

func (d *Deployment) Destroy() {
	for _, c := range d.Clusters {
		c.Destroy()
	}
}

func (d *Deployment) ListNodes() (nodes []*Node) {
	for _, c := range d.Clusters {
		nodes = append(nodes, c.Nodes...)
	}

	return
}

func (d *Deployment) masterIP() string {
	return d.master.ExternalIP
}

func (d *Deployment) myctlChannel() string {
	return d.Myctl.Channel
}

func (d *Deployment) myctlImage() string {
	return d.Myctl.Image
}

func parseDeployment(data []byte) (*Deployment, error) {
	d := &Deployment{}
	if err := yaml.Unmarshal(data, d); err != nil {
		return nil, err
	}

	return d, nil
}

func ParseDeployment(path string) (*Deployment, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return parseDeployment(content)
}
