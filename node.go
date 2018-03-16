package main

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"text/template"
)

type Node struct {
	Pool string

	// Virsh
	Name       string
	ExternalIP string
	InternalIP string
	OS         string
	Docker     string

	//  Chiwen
	Cluster string
	Role    string
}

func (n *Node) Image() string {
	return fmt.Sprintf("%s-%s.qcow2", n.OS, n.Docker)
}

func (n *Node) User() string {
	return "root@" + n.ExternalIP
}

func (n *Node) Qcow2() string {
	return fmt.Sprintf("%s/%s.qcow2", config.DirQcow2, n.Name)
}

func (n *Node) String() string {
	return fmt.Sprintf("%s %s %s %s %s %s %s", n.Name, n.ExternalIP, n.InternalIP, n.OS, n.Docker, n.Cluster, n.Role)
}

func (n *Node) Create() error {
	// $CREATE $node "br0#$nodeIP#255.255.255.0#10.10.1.254#8.8.8.8;br0#$dataIP#255.255.255.0" 8 64 0 $BASE_IMAGES/base-ubuntu16.04-docker17.12.1.qcow2
	const virshFormat = `%s %s "br0#%s#255.255.255.0#10.10.1.254#8.8.8.8;br0#%s#255.255.255.0" 8 64 0 %s/%s`
	bashCmd := fmt.Sprintf(virshFormat, config.Create, n.Name, n.ExternalIP, n.InternalIP, config.DirBaseImages, n.Image())

	_, stderr := Output(exec.Command("/bin/bash", bashCmd))
	if stderr != "" {
		return errors.New(stderr)
	}

	return nil
}

func (n *Node) License() error {
	const templateContent = `
{{.sshPass}} scp {{.pathCWLicense}} $node:/root/
{{.sshPass}} ssh {{.user}} << 'EOF'
	cw_path=/var/lib/docker/volumes/chiwen.config/_data
	test -d $chiwen_config_path || mkdir -p $cw_path
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
		"sshPass":       config.SSHPass,
		"pathCWLicense": config.License,
		"user":          n.User(),
	})

	_, stderr := Output(exec.Command("/bin/bash", tmplBuffer.String()))
	if stderr != "" {
		return errors.New(stderr)
	}

	return nil
}

func (n *Node) Destroy() error {
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
		"qcow2": n.Qcow2(),
	})

	_, stderr := Output(exec.Command("/bin/bash", tmplBuffer.String()))
	if stderr != "" {
		return errors.New(stderr)
	}

	return nil
}

// use elite
func (n *Node) Join() error {
	return nil
}

// use elite
func (n *Node) Init() error {
	return nil
}

// use elite
func (n *Node) Deploy() error {
	return nil
}

func (n *Node) Parse(line string) error {
	list := strings.Split(line, " ")
	if len(list) != 7 {
		return fmt.Errorf("invalid line: %s", line)
	}

	n.Name = list[0]
	n.ExternalIP = list[1]
	n.InternalIP = list[2]
	n.OS = list[3]
	n.Docker = list[4]
	n.Cluster = list[5]
	n.Role = list[6]

	return nil
}
