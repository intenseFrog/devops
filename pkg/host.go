package pkg

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"text/template"

	log "github.com/sirupsen/logrus"
)

const DM = "docker-machine"

type Host struct {
	Name             string   `yaml:"name"`
	ExternalIP       string   `yaml:"external_ip"`
	InternalIP       string   `yaml:"internal_ip"`
	OS               string   `yaml:"os"`
	Docker           string   `yaml:"docker"`
	CPU              *string  `yaml:"cpu,omitempty"`
	Memory           *string  `yaml:"mem,omitempty"`
	Disk             *string  `yaml:"disk,omitempty"`
	InsecureRegistry []string `yaml:"insecure_registry"`
	//  Chiwen
	Role string `yaml:"role"`

	deployment *Deployment
}

func (h *Host) createArgs() (args []string) {
	args = append(args, "create", "-d", "my", "--my-ip", h.ExternalIP, "--my-ip", h.InternalIP)

	if h.CPU != nil {
		args = append(args, "--my-cpu-count", *h.CPU)
	}

	if h.Memory != nil {
		args = append(args, "--my-memory", *h.Memory)
	}

	for _, ir := range h.deployment.InsecureRegistry {
		args = append(args, "--my-insecure-registry", ir)
	}

	for _, ir := range h.InsecureRegistry {
		args = append(args, "--my-insecure-registry", ir)
	}

	args = append(args, h.Name)
	return
}

func (h *Host) Create() error {
	log.Infof("Creating %s...", h.Name)
	// docker-machine create -d my --my-ip 10.10.1.195 --my-insecure-registry 10.10.1.195:5000 luke195
	_, stderr := Output(Exec(DM, h.createArgs()...))
	if stderr != "" {
		if strings.Contains(stderr, "already exists") {
			return nil
		}

		return errors.New(stderr)
	}

	return nil
}

// Deploy myctl
func (h *Host) Deploy() error {
	const templateContent = `
{{.ssh}} << 'EOF'
	docker pull {{.chiwen}}
	id=$(docker create {{.chiwen}})
	docker cp $id:/opt/chiwen/bin/my $HOME/
	$HOME/my \
		deploy \
		--host-ip={{.internalIP}} \
		--domain={{.externalIP}} \
		--registry={{.registry}} \
		{{- range .options}}
		{{.}} \
		{{- end}}
		-y
	docker rm $id
EOF
`
	tmpl, _ := template.New("deploy").Parse(templateContent)
	var b bytes.Buffer
	if err := tmpl.Execute(&b, &map[string]interface{}{
		"ssh":        h.ssh(),
		"chiwen":     h.deployment.chiwenImage(),
		"internalIP": h.InternalIP,
		"externalIP": h.ExternalIP,
		"registry":   h.deployment.registry(),
		"options":    h.deployment.Chiwen.Options,
	}); err != nil {
		return err
	}

	Output(Exec("/bin/bash", "-c", b.String()))

	if web := h.deployment.webImage(); web != "" {
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
			"web": web,
		}); err != nil {
			return err
		}

		Output(Exec("/bin/bash", "-c", webBuffer.String()))
	}

	return nil
}

func (h *Host) Delete() error {
	log.Infof("Deleting %s...", h.Name)
	Output(Exec(DM, "rm", "-y", h.Name))
	return nil
}

func (h *Host) Exist() bool {
	stdout, _ := Output(Exec(DM, "ls", "--filter", fmt.Sprintf("name=%s", h.Name), "-q"))
	return h.Name == stdout
}

func (h *Host) image() string {
	return fmt.Sprintf("%s-%s.qcow2", h.OS, h.Docker)
}

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

func (h *Host) Join() error {
	const sshTemplate = `
{{ .SSH }} << 'EOF'
	{{ .Command }}
EOF
`
	cmd, _ := my("host", "deploy-script", "-q")
	var buf bytes.Buffer
	tmpl, _ := template.New("ssh").Parse(sshTemplate)
	if err := tmpl.Execute(&buf, &map[string]interface{}{
		"SSH":     h.ssh(),
		"Command": cmd,
	}); err != nil {
		return err
	}

	Output(Exec("/bin/bash", "-c", buf.String()))
	return nil
}
