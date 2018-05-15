package common

import (
	"errors"
	"fmt"
	"io/ioutil"
	"sync"

	"gopkg.in/yaml.v2"
)

type Deployment struct {
	Myctl struct {
		Image   string `yaml:"image"`
		Channel string `yaml:"channel"`
		Web     string `yaml:"web"`
	} `yaml:"myctl"`
	Clusters []*Cluster `yaml:"clusters"`

	master *Node
}

func (d *Deployment) setMaster() {
	for i := range d.Clusters {
		cluster := d.Clusters[i]
		cluster.deployment = d

		if master := cluster.Normalize(); master != nil {
			d.master = master
		}
	}
}

func (d *Deployment) License() error {
	if d.master == nil {
		return errors.New("master not set, skip licensing")
	}

	fmt.Println("Licensing...")
	return d.master.License()
}

func (d *Deployment) CleanKnownHosts() {
	for _, c := range d.Clusters {
		c.CleanKnownHosts()
	}
}

func (d *Deployment) Create() error {
	var wg sync.WaitGroup
	wg.Add(len(d.ListNodes()))

	for _, n := range d.ListNodes() {
		go func(n *Node) {
			defer wg.Done()
			if err := n.Create(); err != nil {
				panic(err)
			}
		}(n)
	}
	wg.Wait()

	if err := d.License(); err != nil {
		fmt.Printf("Failed licensing: %s\n", err.Error())
	}

	return nil
}

func (d *Deployment) Update() error {
	fmt.Println("Updating master...")
	return d.master.Deploy()
}

func (d *Deployment) Deploy() (err error) {
	defer eliteLogout()

	if err = d.License(); err != nil {
		fmt.Printf("Failed licensing: %s\n", err.Error())
	}

	fmt.Println("Deploying master...")
	if err = d.master.Deploy(); err != nil {
		return err
	}

	eliteLogin(d.master.ExternalIP)

	fmt.Println("Deploying clusters...")

	var wg sync.WaitGroup
	wg.Add(len(d.Clusters))

	for _, cluster := range d.Clusters {
		go func(c *Cluster) {
			defer wg.Done()
			if err = c.Deploy(); err != nil {
				panic(err)
			}
		}(cluster)
	}

	wg.Wait()
	return
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

func (d *Deployment) myctlWeb() string {
	return d.Myctl.Web
}

func parseDeployment(data []byte) (*Deployment, error) {
	d := &Deployment{}
	if err := yaml.Unmarshal(data, d); err != nil {
		return nil, err
	}

	d.setMaster()
	return d, nil
}

func ParseDeployment(path string) (*Deployment, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return parseDeployment(content)
}
