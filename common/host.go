package common

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"text/template"
)

const DM = "docker-machine"

type Host struct {
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

	deployment *Deployment
	options    []string
}

func (h *Host) createArgs() (args []string) {
	args = append(args, "create", "-d", "my", "--my-ip", h.ExternalIP, "--my-ip", h.InternalIP)

	if h.CPU != nil {
		args = append(args, "--my-cpu-count", *h.CPU)
	}

	if h.Memory != nil {
		args = append(args, "--my-memory", *h.Memory)
	}

	for _, ir := range h.deployment.InsecureRegistries {
		args = append(args, "--engine-insecure-registry", ir)
	}

	args = append(args, h.Name)
	return
}

func (h *Host) Create() error {
	fmt.Printf("Creating %s...\n", h.Name)
	// docker-machine create -d my --my-ip 10.10.1.195 --engine-insecure-registry 10.10.1.195:5000 luke195
	_, stderr := Output(exec.Command(DM, h.createArgs()...))
	if stderr != "" {
		return errors.New(stderr)
	}

	return nil
}

// Deploy myctl
func (h *Host) Deploy() error {
	const templateContent = `
{{.ssh}} << 'EOF'
	docker pull {{.myctl}}
	docker run \
		--rm \
		--net=host \
		--cap-add=NET_ADMIN \
		-e MYCTL_IMAGE={{.myctl}} \
		-e SSH_PORT=${SSH_CLIENT##* } \
		-e SSH_USER=$(id -un) \
		-e SSH_USER_HOME=$HOME \
		-w /rootfs/$(pwd) \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-v /:/rootfs \
		-v chiwen.config:/etc/chiwen \
		{{.myctl}} deploy \
		--advertise-ip={{.internalIP}} \
		--domain={{.externalIP}} \
		{{- range .options}}
		{{.}} \
		{{- end}}
		-y
EOF
`
	tmplDeploy, _ := template.New("deploy").Parse(templateContent)
	var tmplBuffer bytes.Buffer
	if err := tmplDeploy.Execute(&tmplBuffer, &map[string]interface{}{
		"ssh":        h.ssh(),
		"myctl":      h.deployment.myctlImage(),
		"internalIP": h.InternalIP,
		"externalIP": h.ExternalIP,
		"options":    h.options,
	}); err != nil {
		return err
	}

	Output(exec.Command("/bin/bash", "-c", tmplBuffer.String()))

	if web := h.deployment.myctlWeb(); web != "" {
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
			"ssh": h.ssh(),
			"web": h.deployment.myctlWeb(),
		}); err != nil {
			return err
		}

		Output(exec.Command("/bin/bash", "-c", webBuffer.String()))
	}

	return nil
}

func (h *Host) Destroy() error {
	fmt.Printf("Destroying %s...\n", h.Name)
	Output(exec.Command(DM, "rm", "-y", h.Name))
	return nil
}

func (h *Host) Exist() bool {
	stdout, _ := Output(exec.Command(DM, "ls", "--filter", fmt.Sprintf("name=%s", h.Name), "-q"))
	return h.Name == stdout
}

func (h *Host) image() string {
	return fmt.Sprintf("%s-%s.qcow2", h.OS, h.Docker)
}

// func (h *Host) qcow2() string {
// 	return fmt.Sprintf("%s/%s.qcow2", config.DirQcow2, h.Name)
// }

func (h *Host) scp(src, dst string) string {
	// return fmt.Sprintf("scp -o StrictHostKeyChecking=no %s %s", src, dst)
	return fmt.Sprintf("%s scp -r %s %s", DM, src, dst)
}

func (h *Host) ssh() string {
	return fmt.Sprintf("%s ssh %s", DM, h.Name)
}

func (h *Host) userAtHost() string {
	return "root@" + h.Name
}

const sshTemplate = `
{{ .SSH }} << 'EOF'
	{{ .Command }}
EOF
`

func (h *Host) Join() error {
	cmd, _ := elite("host", "deploy-script", "-q")
	var buf bytes.Buffer
	tmpl, _ := template.New("ssh").Parse(sshTemplate)
	if err := tmpl.Execute(&buf, &map[string]interface{}{
		"SSH":     h.ssh(),
		"Command": cmd,
	}); err != nil {
		return err
	}

	Output(exec.Command("/bin/bash", "-c", buf.String()))
	return nil
}
