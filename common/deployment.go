package common

import (
	"fmt"
	"io/ioutil"
	"sync"
	"time"

	yaml "gopkg.in/yaml.v2"
)

type Deployment struct {
	Myctl struct {
		Image   string   `yaml:"image"`
		Web     string   `yaml:"web"`
		Options []string `yaml:"options"`
	} `yaml:"myctl"`
	Master             *Host      `yaml:"master"`
	Hosts              []*Host    `yaml:"hosts"`
	Clusters           []*Cluster `yaml:"clusters"`
	InsecureRegistries []string   `yaml:"insecure-registry"`
}

func (d *Deployment) Create() error {
	var wg sync.WaitGroup
	hosts := d.ListHosts()
	wg.Add(len(hosts))

	for _, h := range hosts {
		go func(h *Host) {
			defer wg.Done()
			if !h.Exist() {
				fmt.Println(h.Name)
				if err := h.Create(); err != nil {
					panic(err)
				}
			}
		}(h)
	}
	wg.Wait()

	return nil
}

func (d *Deployment) Update() error {
	fmt.Println("Updating master...")
	return d.Master.Deploy()
}

func (d *Deployment) Deploy() (err error) {
	defer eliteLogout()

	fmt.Println("Deploying master...")
	if err = d.Master.Deploy(); err != nil {
		return err
	}

	// try to login
	// also works as a health-check to see if chiwen is ready
	eliteLogin(d.Master.ExternalIP, 5*time.Minute)

	var wg sync.WaitGroup
	wg.Add(len(d.Hosts))
	fmt.Println("Joining hosts...")
	for i := range d.Hosts {
		h := d.Hosts[i]
		go func() {
			defer wg.Done()
			if err = h.Join(); err != nil {
				panic(err)
			}
		}()
	}
	wg.Wait()

	fmt.Println("Deploying clusters...")
	wg.Add(len(d.Clusters))
	for i := range d.Clusters {
		c := d.Clusters[i]
		go func() {
			defer wg.Done()
			c.Deploy()
		}()
	}
	wg.Wait()

	return
}

func (d *Deployment) Destroy() {
	var wg sync.WaitGroup
	wg.Add(len(d.ListHosts()))

	for _, h := range d.ListHosts() {
		go func(h *Host) {
			defer wg.Done()
			h.Destroy()
		}(h)
	}

	wg.Wait()
}

func (d *Deployment) ListHosts() (hosts []*Host) {
	if d.Master != nil {
		hosts = append(hosts, d.Master)
	}
	hosts = append(hosts, d.Hosts...)
	return hosts
}

func (d *Deployment) masterIP() string {
	return d.Master.ExternalIP
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

	// set deployment
	if d.Master != nil {
		d.Master.deployment = d
	}

	for i := range d.Hosts {
		d.Hosts[i].deployment = d
	}
	for i := range d.Clusters {
		d.Clusters[i].deployment = d
		d.Clusters[i].Normalize()
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
