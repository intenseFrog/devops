package common

import (
	"bytes"
	"fmt"
	"os/exec"

	"text/template"
)

type clusterNode interface {
	init() error
	join() error
}

type swarmNode struct {
	infraNode *Node
}

func (n *swarmNode) init() error {
	const templateContent = `
{{.sshPass}} {{.ssh}} << 'EOF'
	{{.deployCmd}}
EOF
	`

	node := n.infraNode

	elite("login", "-u", "admin", "-p", "admin", node.masterIP())

	createArgs := []string{"cluster", "create", node.clusterName(), "--swarm"}
	for k, v := range node.cluster.Params {
		createArgs = append(createArgs, "-p", fmt.Sprintf("%s=%s", k, v))
	}
	elite(createArgs...)

	elite("cluster", "use", node.clusterName())

	deployScript := elite("node", "deploy-script", "-q", fmt.Sprintf("--ip=%s", node.InternalIP))

	var tmplBuffer bytes.Buffer
	tmplDeploy, _ := template.New("deploy-script").Parse(templateContent)
	tmplDeploy.Execute(&tmplBuffer, &map[string]interface{}{
		"sshPass":   config.SSHPass,
		"ssh":       node.ssh(),
		"deployCmd": deployScript,
	})

	_, stderr := Output(exec.Command("/bin/bash", "-c", tmplBuffer.String()))
	if stderr != "" {
		fmt.Println(stderr)
	}

	return nil
}

func (n *swarmNode) join() error {
	const templateContent = `
{{.sshPass}} {{.ssh}} << 'EOF'
	{{.deployCmd}}
EOF
`

	deployScript := elite("node", "deploy-script", "-q")

	var tmplBuffer bytes.Buffer
	tmplDeploy, _ := template.New("deploy-script").Parse(templateContent)
	tmplDeploy.Execute(&tmplBuffer, &map[string]interface{}{
		"sshPass":   config.SSHPass,
		"ssh":       n.infraNode.ssh(),
		"deployCmd": deployScript,
	})

	_, stderr := Output(exec.Command("/bin/bash", "-c", tmplBuffer.String()))
	if stderr != "" {
		fmt.Println(stderr)
	}

	return nil
}

type kubernetesNode struct {
	infraNode *Node
}

func (n *kubernetesNode) init() error {
	const templateContent = `
{{.sshPass}} {{.ssh}} << 'EOF'
	{{.deployCmd}}
EOF
	`

	node := n.infraNode

	elite("login", "-u", "admin", "-p", "admin", node.masterIP())

	createArgs := []string{"cluster", "create", node.clusterName(), "--kubernetes"}
	for k, v := range node.cluster.Params {
		createArgs = append(createArgs, "-p", fmt.Sprintf("%s=%s", k, v))
	}
	elite(createArgs...)

	elite("cluster", "use", node.clusterName())

	deployScript := elite("node", "deploy-script", "-q", fmt.Sprintf("--ip=%s", node.InternalIP))

	var tmplBuffer bytes.Buffer
	tmplDeploy, _ := template.New("deploy-script").Parse(templateContent)
	tmplDeploy.Execute(&tmplBuffer, &map[string]interface{}{
		"sshPass":   config.SSHPass,
		"ssh":       node.ssh(),
		"deployCmd": deployScript,
	})

	_, stderr := Output(exec.Command("/bin/bash", "-c", tmplBuffer.String()))
	if stderr != "" {
		fmt.Println(stderr)
	}

	return nil
}

func (n *kubernetesNode) join() error {
	const templateContent = `
{{.sshPass}} {{.ssh}} << 'EOF'
	{{.deployCmd}}
EOF
`

	node := n.infraNode

	deployScript := elite("node", "deploy-script", "-q", fmt.Sprintf("--ip=%s", node.InternalIP))

	tmplDeploy, _ := template.New("deploy-script").Parse(templateContent)
	var tmplBuffer bytes.Buffer
	tmplDeploy.Execute(&tmplBuffer, &map[string]interface{}{
		"sshPass":   config.SSHPass,
		"ssh":       n.infraNode.ssh(),
		"deployCmd": deployScript,
	})

	_, stderr := Output(exec.Command("/bin/bash", "-c", tmplBuffer.String()))
	if stderr != "" {
		fmt.Println(stderr)
	}

	return nil
}
