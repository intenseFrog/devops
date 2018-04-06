package common

import (
	"bytes"
	"fmt"
	"os/exec"

	"text/template"
)

const templateDeploy = `
{{.sshPass}} {{.ssh}} << 'EOF'
	{{.deployCmd}}
EOF
`

type clusterNode interface {
	init() error
	join() error
}

type swarmNode struct {
	infraNode *Node
}

func (n *swarmNode) init() error {
	node := n.infraNode

	createArgs := []string{"cluster", "create", node.clusterName(), "--swarm"}
	for k, v := range node.cluster.Params {
		createArgs = append(createArgs, "-p", fmt.Sprintf("%s=%s", k, v))
	}

	eliteArgs := &EliteArguments{}

	eliteArgs.Append(false, createArgs...)
	eliteArgs.Append(false, "cluster", "use", node.clusterName())
	eliteArgs.Append(true, "node", "deploy-script", "-q", fmt.Sprintf("--ip=%s", node.InternalIP))

	deployScript := elite(eliteArgs)[0]

	var tmplBuffer bytes.Buffer
	tmplDeploy, _ := template.New("deploy-script").Parse(templateDeploy)
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
	node := n.infraNode

	createArgs := []string{"cluster", "create", node.clusterName(), "--swarm"}
	for k, v := range node.cluster.Params {
		createArgs = append(createArgs, "-p", fmt.Sprintf("%s=%s", k, v))
	}

	eliteArgs := &EliteArguments{}

	eliteArgs.Append(false, createArgs...)
	eliteArgs.Append(false, "cluster", "use", node.clusterName())
	eliteArgs.Append(true, "node", "deploy-script", "-q")

	deployScript := elite(eliteArgs)[0]

	var tmplBuffer bytes.Buffer
	tmplDeploy, _ := template.New("deploy-script").Parse(templateDeploy)
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

type kubernetesNode struct {
	infraNode *Node
}

func (n *kubernetesNode) init() error {
	node := n.infraNode

	createArgs := []string{"cluster", "create", node.clusterName(), "--kubernetes"}
	for k, v := range node.cluster.Params {
		createArgs = append(createArgs, "-p", fmt.Sprintf("%s=%s", k, v))
	}

	eliteArgs := &EliteArguments{}

	eliteArgs.Append(false, createArgs...)
	eliteArgs.Append(false, "cluster", "use", node.clusterName())
	eliteArgs.Append(true, "node", "deploy-script", "-q", fmt.Sprintf("--ip=%s", node.InternalIP))

	deployScript := elite(eliteArgs)[0]

	var tmplBuffer bytes.Buffer
	tmplDeploy, _ := template.New("deploy-script").Parse(templateDeploy)
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
	node := n.infraNode

	eliteArgs := &EliteArguments{}

	eliteArgs.Append(false, "cluster", "use", node.clusterName())
	eliteArgs.Append(true, "node", "deploy-script", "-q", fmt.Sprintf("--ip=%s", node.InternalIP))

	deployScript := elite(eliteArgs)[0]

	tmplDeploy, _ := template.New("deploy-script").Parse(templateDeploy)
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
