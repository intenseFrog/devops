package common

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"

	"text/template"
)

const DM = "docker-machine"

type Node struct {
	Name       string  `yaml:"name"`
	ExternalIP string  `yaml:"external_ip"`
	InternalIP string  `yaml:"internal_ip"`
	OS         string  `yaml:"os"`
	Docker     string  `yaml:"docker"`
	CPU        *string `yaml:"cpu,omitempty"`
	Memory     *string `yaml:"mem,omitempty"`
	Disk       *string `yaml:"disk,omitempty"`
	//  Chiwen
	Role string `yaml:"role"`

	cluster *Cluster
}

func (n *Node) clusterName() string {
	return n.cluster.Name
}

func (n *Node) clusterKind() string {
	return n.cluster.Kind
}

func (n *Node) CleanKnownHost() {
	// Output(exec.Command("ssh-keygen", "-f", "/root/.ssh/known_hosts", "-R", n.ExternalIP))
	Output(exec.Command("ssh-keygen", "-R", n.Name))
}

func (n *Node) createArgs() (args []string) {
	args = append(args, "create", "-d", "my", "--my-ip", n.ExternalIP, "--my-ip", n.InternalIP)

	if n.CPU != nil {
		args = append(args, "--my-cpu-count", *n.CPU)
	}

	if n.Memory != nil {
		args = append(args, "--my-memory", *n.Memory)
	}

	for _, ir := range n.cluster.deployment.InsecureRegistries {
		args = append(args, "--engine-insecure-registry", ir)
	}

	args = append(args, n.Name)
	return
}

func (n *Node) Create() error {
	fmt.Printf("Creating %s...\n", n.Name)
	// docker-machine create -d my --my-ip 10.10.1.195 --engine-insecure-registry 10.10.1.195:5000 luke195
	_, stderr := Output(exec.Command(DM, n.createArgs()...))
	if stderr != "" {
		return errors.New(stderr)
	}

	return nil
}

func (n *Node) Deploy() error {
	const templateContent = `
{{.ssh}} << 'EOF'
	docker pull {{.myctl}}
	docker run --rm --net=host \
    	-v /var/run/docker.sock:/var/run/docker.sock \
		-v chiwen.config:/etc/chiwen \
		-e MYCTL_IMAGE={{.myctl}} \
		{{.myctl}} deploy \
		--advertise-ip={{.internalIP}} \
		--domain={{.externalIP}} \
		-y
EOF
`
	tmplDeploy, _ := template.New("deploy").Parse(templateContent)
	var tmplBuffer bytes.Buffer
	if err := tmplDeploy.Execute(&tmplBuffer, &map[string]interface{}{
		"ssh":        n.ssh(),
		"myctl":      n.cluster.myctlImage(),
		"internalIP": n.InternalIP,
		"externalIP": n.ExternalIP,
	}); err != nil {
		return err
	}

	Output(exec.Command("/bin/bash", "-c", tmplBuffer.String()))

	if web := n.cluster.myctlWeb(); web != "" {
		const webTemplate = `
{{.ssh}} << 'EOF'
	docker pull {{.web}}
	docker run \
		-v chiwen.web:/data \
		{{.web}}
EOF
`
		tmplWeb, _ := template.New("web").Parse(webTemplate)
		var webBuffer bytes.Buffer
		if err := tmplWeb.Execute(&webBuffer, &map[string]interface{}{
			"ssh": n.ssh(),
			"web": n.cluster.myctlWeb(),
		}); err != nil {
			return err
		}

		Output(exec.Command("/bin/bash", "-c", webBuffer.String()))
	}

	return nil
}

func (n *Node) Destroy() error {
	fmt.Printf("Destroying %s...\n", n.Name)
	Output(exec.Command(DM, "rm", "-y", n.Name))
	return nil
}

func (n *Node) Exist() bool {
	stdout, _ := Output(exec.Command(DM, "ls", "--filter", fmt.Sprintf("name=%s", n.Name), "-q"))
	return n.Name == stdout
}

func (n *Node) image() string {
	return fmt.Sprintf("%s-%s.qcow2", n.OS, n.Docker)
}

func (n *Node) kubernetesNode() clusterNode {
	return &kubernetesNode{infraNode: n}
}

func (n *Node) License() error {
	const templateContent = `
{{.scp}}
{{.ssh}} << 'EOF'
	cw_path=/var/lib/docker/volumes/chiwen.config/_data
	test -d $cw_path || mkdir -p $cw_path
	mac=$(cat /sys/class/net/$(ip route show default|awk '/default/ {print $5}')/address)
	hw_sig=$(echo -n "${mac}HJLXZZ" | openssl dgst -md5 -binary | openssl enc -base64)
	/root/chiwen-license \
		-id dummy \
		-ed PE \
		-hw $hw_sig \
		-ia $(date -u +“%Y-%m-%d”) \
		-ib minhao.jin \
		-ea 2049-12-31 \
		-o chiwen-team \
		-p features=Megabric > $cw_path/license.key
EOF
`

	tmplLicense, _ := template.New("license").Parse(templateContent)
	var tmplBuffer bytes.Buffer
	if err := tmplLicense.Execute(&tmplBuffer, &map[string]interface{}{
		"ssh": n.ssh(),
		"scp": n.scp(config.License, fmt.Sprintf("%s:/root/", n.userAtNode())),
	}); err != nil {
		return err
	}

	Output(exec.Command("/bin/bash", "-c", tmplBuffer.String()))
	return nil
}

func (n *Node) masterIP() string {
	return n.cluster.masterIP()
}

// func (n *Node) qcow2() string {
// 	return fmt.Sprintf("%s/%s.qcow2", config.DirQcow2, n.Name)
// }

func (n *Node) scp(src, dst string) string {
	// return fmt.Sprintf("scp -o StrictHostKeyChecking=no %s %s", src, dst)
	return fmt.Sprintf("%s scp -r %s %s", DM, src, dst)
}

func (n *Node) ssh() string {
	return fmt.Sprintf("%s ssh %s", DM, n.Name)
}

func (n *Node) String() string {
	return fmt.Sprintf("%s %s %s %s %s %s %s, %s", n.Name, n.ExternalIP, n.InternalIP, n.OS, n.Docker, n.cluster.Name, n.Role, n.cluster.Params)
}

func (n *Node) swarmNode() clusterNode {
	return &swarmNode{infraNode: n}
}

func (n *Node) userAtNode() string {
	return "root@" + n.Name
}
