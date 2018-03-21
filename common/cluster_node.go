package common

import (
	"bytes"
	"fmt"
	"os/exec"

	"text/template"
)

type ClusterNode interface {
	Init() error
	Join() error
}

type SwarmNode struct {
	InfraNode *Node
}

func (n *SwarmNode) Init() error {
	const templateContent = `
{{.sshPass}} {{.ssh}} << 'EOF'
	{{.deployCmd}}
EOF
	`

	node := n.InfraNode

	Elite("login", "-u", "admin", "-p", "admin", node.deployment.master.ExternalIP)

	createArgs := []string{"cluster", "create", node.Cluster, "--swarm"}
	for k, v := range node.Params {
		createArgs = append(createArgs, "-p", fmt.Sprintf("%s=%s", k, v))
	}
	Elite(createArgs...)

	Elite("cluster", "use", node.Cluster)

	deployScript := Elite("node", "deploy-script", "-q", fmt.Sprintf("--ip=%s", node.InternalIP))

	var tmplBuffer bytes.Buffer
	tmplDeploy, _ := template.New("deploy-script").Parse(templateContent)
	tmplDeploy.Execute(&tmplBuffer, &map[string]interface{}{
		"sshPass":   config.SSHPass,
		"ssh":       node.SSH(),
		"deployCmd": deployScript,
	})

	_, stderr := Output(exec.Command("/bin/bash", "-c", tmplBuffer.String()))
	if stderr != "" {
		fmt.Println(stderr)
	}

	return nil
}

func (n *SwarmNode) Join() error {
	const templateContent = `
{{.sshPass}} {{.ssh}} << 'EOF'
	{{.deployCmd}}
EOF
`

	deployScript := Elite("node", "deploy-script", "-q")

	var tmplBuffer bytes.Buffer
	tmplDeploy, _ := template.New("deploy-script").Parse(templateContent)
	tmplDeploy.Execute(&tmplBuffer, &map[string]interface{}{
		"sshPass":   config.SSHPass,
		"ssh":       n.InfraNode.SSH(),
		"deployCmd": deployScript,
	})

	_, stderr := Output(exec.Command("/bin/bash", "-c", tmplBuffer.String()))
	if stderr != "" {
		fmt.Println(stderr)
	}

	return nil
}

type KubernetesNode struct {
	InfraNode *Node
}

func (n *KubernetesNode) Init() error {
	const templateContent = `
{{.sshPass}} {{.ssh}} << 'EOF'
	{{.deployCmd}}
EOF
	`

	node := n.InfraNode

	Elite("login", "-u", "admin", "-p", "admin", node.deployment.master.ExternalIP)

	createArgs := []string{"cluster", "create", node.Cluster, "--kubernetes"}
	for k, v := range node.Params {
		createArgs = append(createArgs, "-p", fmt.Sprintf("%s=%s", k, v))
	}
	Elite(createArgs...)

	Elite("cluster", "use", node.Cluster)

	deployScript := Elite("node", "deploy-script", "-q", fmt.Sprintf("--ip=%s", node.InternalIP))

	var tmplBuffer bytes.Buffer
	tmplDeploy, _ := template.New("deploy-script").Parse(templateContent)
	tmplDeploy.Execute(&tmplBuffer, &map[string]interface{}{
		"sshPass":   config.SSHPass,
		"ssh":       node.SSH(),
		"deployCmd": deployScript,
	})

	_, stderr := Output(exec.Command("/bin/bash", "-c", tmplBuffer.String()))
	if stderr != "" {
		fmt.Println(stderr)
	}

	return nil
}

func (n *KubernetesNode) Join() error {
	const templateContent = `
{{.sshPass}} {{.ssh}} << 'EOF'
	{{.deployCmd}}
EOF
`

	deployScript := Elite("node", "deploy-script", "-q", fmt.Sprintf("--ip=%s", n.InfraNode.InternalIP))

	tmplDeploy, _ := template.New("deploy-script").Parse(templateContent)
	var tmplBuffer bytes.Buffer
	tmplDeploy.Execute(&tmplBuffer, &map[string]interface{}{
		"sshPass":   config.SSHPass,
		"ssh":       n.InfraNode.SSH(),
		"deployCmd": deployScript,
	})

	_, stderr := Output(exec.Command("/bin/bash", "-c", tmplBuffer.String()))
	if stderr != "" {
		fmt.Println(stderr)
	}

	return nil
}
