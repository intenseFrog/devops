package common

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"

	"text/template"
)

type Node struct {
	// Virsh
	Name       string `yaml:"name"`
	ExternalIP string `yaml:"external_ip"`
	InternalIP string `yaml:"internal_ip"`
	OS         string `yaml:"os"`
	Docker     string `yaml:"docker"`

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
	// out, stderr := Output(exec.Command("ssh-keygen", "-f", "/root/.ssh/known_hosts", "-R", n.ExternalIP))
	out, stderr := Output(exec.Command("ssh-keygen", "-R", n.Name))
	if stderr != "" {
		fmt.Println(stderr)
	}

	fmt.Println(out)
}

func (n *Node) Create() error {
	fmt.Printf("Creating %s...\n", n.Name)
	// /devops/create_vms_2d.sh developer183 "br0#10.10.1.183#255.255.255.0#10.10.1.254#8.8.8.8;br0#172.16.88.183#255.255.255.0" 8 64 0 /devops/base_images/ubuntu16.04-docker17.12.1.qcow2
	network := fmt.Sprintf("br0#%s#255.255.255.0#10.10.1.254#8.8.8.8;br0#%s#255.255.255.0", n.ExternalIP, n.InternalIP)
	cpu, memory, disk := "4", "16", "0"
	imagePath := fmt.Sprintf("%s/%s", config.DirBaseImages, n.image())

	out, stderr := Output(exec.Command(config.Create, n.Name, network, cpu, memory, disk, imagePath))
	if stderr != "" {
		return errors.New(stderr)
	}

	fmt.Println(out)
	return nil
}

func (n *Node) Deploy() error {
	const templateContent = `
{{.sshPass}} {{.ssh}} << 'EOF'
	docker pull {{.myctl}}
	docker run --rm --net=host \
    	-v /var/run/docker.sock:/var/run/docker.sock \
    	-v chiwen.config:/etc/chiwen \
		{{.myctl}} deploy \
		-c {{.channel}} \
		--advertise-ip={{.internalIP}} \
		--domain={{.externalIP}} \
		--registry-external={{.externalIP}}
EOF
`
	tmplDeploy, _ := template.New("deploy").Parse(templateContent)
	var tmplBuffer bytes.Buffer
	tmplDeploy.Execute(&tmplBuffer, &map[string]interface{}{
		"sshPass":    config.SSHPass,
		"ssh":        n.ssh(),
		"myctl":      n.cluster.myctlImage(),
		"channel":    n.cluster.myctlChannel(),
		"internalIP": n.InternalIP,
		"externalIP": n.ExternalIP,
	})

	_, stderr := Output(exec.Command("/bin/bash", "-c", tmplBuffer.String()))
	if stderr != "" {
		fmt.Println(stderr)
	}

	if web := n.cluster.myctlWeb(); web != "" {
		const webTemplate = `
{{.sshPass}} {{.ssh}} << 'EOF'
	docker pull {{.web}}
	docker run \
		-v chiwen.web:/data \
		{{.web}}
EOF
`
		tmplWeb, _ := template.New("web").Parse(webTemplate)
		var webBuffer bytes.Buffer
		tmplWeb.Execute(&webBuffer, &map[string]interface{}{
			"sshPass": config.SSHPass,
			"ssh":     n.ssh(),
			"web":     n.cluster.myctlWeb(),
		})

		_, stderr := Output(exec.Command("/bin/bash", "-c", webBuffer.String()))
		if stderr != "" {
			fmt.Println(stderr)
		}
	}

	return nil
}

func (n *Node) Destroy() error {
	fmt.Printf("Destroying %s...\n", n.Name)
	const templateContent = `
virsh destroy {{.name}}
virsh undefine {{.name}}
if [ -e "{{.qcow2}}" ]; then
	rm {{.qcow2}}
fi
`
	tmplDestroy, _ := template.New("destroy").Parse(templateContent)
	var tmplBuffer bytes.Buffer
	tmplDestroy.Execute(&tmplBuffer, &map[string]interface{}{
		"name":  n.Name,
		"qcow2": n.qcow2(),
	})

	_, stderr := Output(exec.Command("/bin/bash", "-c", tmplBuffer.String()))
	if stderr != "" {
		return errors.New(stderr)
	}

	return nil
}

func (n *Node) image() string {
	return fmt.Sprintf("%s-%s.qcow2", n.OS, n.Docker)
}

func (n *Node) kubernetesNode() clusterNode {
	return &kubernetesNode{infraNode: n}
}

func (n *Node) License() error {
	const templateContent = `
{{.sshPass}} {{.scp}}
{{.sshPass}} {{.ssh}} << 'EOF'
	cw_path=/var/lib/docker/volumes/chiwen.config/_data
	test -d $cw_path || mkdir -p $cw_path
	mac=$(cat /sys/class/net/$(ip route show default|awk '/default/ {print $5}')/address)
	hw_sig=$(echo -n "${mac}HJLXZZ" | openssl dgst -md5 -binary | openssl enc -base64)
	/root/chiwen-license \
		-id dummy \
		-hw $hw_sig \
		-ia $(date -u +“%Y-%m-%d”) \
		-ib minhao.jin \
		-ea 2049-12-31 \
		-o devops@160 > $cw_path/license.key
EOF
`

	tmplLicense, _ := template.New("license").Parse(templateContent)
	var tmplBuffer bytes.Buffer
	tmplLicense.Execute(&tmplBuffer, &map[string]interface{}{
		"sshPass": config.SSHPass,
		"ssh":     n.ssh(),
		"scp":     n.scp(config.License, fmt.Sprintf("%s:/root/", n.userAtNode())),
	})

	_, stderr := Output(exec.Command("/bin/bash", "-c", tmplBuffer.String()))
	if stderr != "" {
		fmt.Println(stderr)
	}

	return nil
}

func (n *Node) masterIP() string {
	return n.cluster.masterIP()
}

func (n *Node) qcow2() string {
	return fmt.Sprintf("%s/%s.qcow2", config.DirQcow2, n.Name)
}

func (n *Node) scp(src, dst string) string {
	return fmt.Sprintf("scp -o StrictHostKeyChecking=no %s %s", src, dst)
}

func (n *Node) ssh() string {
	return "ssh -o StrictHostKeyChecking=no root@" + n.ExternalIP
}

func (n *Node) String() string {
	return fmt.Sprintf("%s %s %s %s %s %s %s, %s", n.Name, n.ExternalIP, n.InternalIP, n.OS, n.Docker, n.cluster.Name, n.Role, n.cluster.Params)
}

func (n *Node) swarmNode() clusterNode {
	return &swarmNode{infraNode: n}
}

func (n *Node) userAtNode() string {
	return "root@" + n.ExternalIP
}
