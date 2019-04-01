package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"mydevops/common"
)

// CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o mydevops

func main() {
	RootCmd := &cobra.Command{
		Use:   "mydevops",
		Short: "CLI tool to manage miaoyun",
		Long:  "CLI tool to manage miaoyun",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	applyCmd := &cobra.Command{
		Use:   "apply",
		Short: "create machines and deploy chiwen",
		Long:  "create machines and deploy chiwen",
		RunE:  runApply,
	}
	applyCmd.Flags().Bool("force", false, "destroy previous machines")
	applyCmd.Flags().StringP("file", "f", "", "Specify the file path")
	applyCmd.MarkFlagRequired("file")

	createCmd := &cobra.Command{
		Use:   "create",
		Short: "create a bunch of machines",
		Long:  "create a bunch of machines",
		RunE:  runCreate,
	}
	createCmd.Flags().Bool("force", false, "destroy previous machines")
	createCmd.Flags().StringP("file", "f", "", "Specify the file path")
	createCmd.MarkFlagRequired("file")

	deployCmd := &cobra.Command{
		Use:   "deploy",
		Short: "create and deploy a cluster",
		Long:  "create and deploy a cluster",
		RunE:  runDeploy,
	}
	deployCmd.Flags().StringP("file", "f", "", "Specify the file path")
	deployCmd.MarkFlagRequired("file")

	destroyCmd := &cobra.Command{
		Use:   "destroy",
		Short: "destroy nodes defined by yaml file",
		Long:  "destroy nodes defined by yaml file",
		RunE:  runDestroy,
	}
	destroyCmd.Flags().StringP("file", "f", "", "Specify the file path")
	destroyCmd.Flags().Bool("force", false, "force destroying machines")

	exampleCmd := &cobra.Command{
		Use:   "example",
		Short: "print out an example of a yaml file",
		Long:  "print out an example of a yaml file",
		Run:   runExample,
	}

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "list nodes",
		Long:  "list nodes",
		RunE:  runList,
	}
	listCmd.Flags().BoolP("quiet", "q", false, "List names only")

	parseCmd := &cobra.Command{
		Use:    "parse",
		Hidden: true,
		RunE:   runParse,
	}
	parseCmd.Flags().StringP("file", "f", "", "Specify the file path")

	updateCmd := &cobra.Command{
		Use:   "update",
		Short: "update a cluster",
		Long:  "update a cluster",
		RunE:  runUpdate,
	}
	updateCmd.Flags().StringP("file", "f", "", "Specify the file path")
	updateCmd.MarkFlagRequired("file")

	RootCmd.AddCommand(applyCmd)
	RootCmd.AddCommand(createCmd)
	RootCmd.AddCommand(deployCmd)
	RootCmd.AddCommand(exampleCmd)
	RootCmd.AddCommand(listCmd)
	RootCmd.AddCommand(destroyCmd)
	RootCmd.AddCommand(updateCmd)

	// back door
	RootCmd.AddCommand(parseCmd)

	if err := RootCmd.Execute(); err != nil {
		os.Exit(-1)
	}
}

func runApply(cmd *cobra.Command, args []string) error {
	start := time.Now()
	defer common.PrintDone(start)

	path, err := cmd.Flags().GetString("file")
	if err != nil {
		return err
	}

	deploy, err := common.ParseDeployment(path)
	if err != nil {
		return err
	}

	if force, _ := cmd.Flags().GetBool("force"); force {
		deploy.Destroy()
	}

	if err = deploy.Create(); err != nil {
		return err
	}

	return deploy.Deploy()
}

func runCreate(cmd *cobra.Command, args []string) error {
	start := time.Now()
	defer common.PrintDone(start)

	path, err := cmd.Flags().GetString("file")
	if err != nil {
		return err
	}

	deploy, err := common.ParseDeployment(path)
	if err != nil {
		return err
	}

	if force, _ := cmd.Flags().GetBool("force"); force {
		deploy.Destroy()
	}

	return deploy.Create()
}

func runDeploy(cmd *cobra.Command, args []string) error {
	start := time.Now()
	defer common.PrintDone(start)

	path, err := cmd.Flags().GetString("file")
	if err != nil {
		return err
	}

	deploy, err := common.ParseDeployment(path)
	if err != nil {
		return err
	}

	return deploy.Deploy()
}

func runDestroy(cmd *cobra.Command, args []string) error {
	path, err := cmd.Flags().GetString("file")
	if err != nil {
		return err
	}

	deploy, err := common.ParseDeployment(path)
	if err != nil {
		return err
	}

	var names []string
	for _, h := range deploy.ListHosts() {
		names = append(names, h.Name)
	}

	force, _ := cmd.Flags().GetBool("force")
	msg := fmt.Sprintf("About to remove %s", strings.Join(names, ", "))

	if force || common.Confirm(msg) {
		deploy.Destroy()
	}

	return nil
}

func runUpdate(cmd *cobra.Command, args []string) error {
	start := time.Now()
	defer common.PrintDone(start)

	path, err := cmd.Flags().GetString("file")
	if err != nil {
		return err
	}

	deploy, err := common.ParseDeployment(path)
	if err != nil {
		return err
	}

	return deploy.Update()
}

func runList(cmd *cobra.Command, args []string) error {
	quiet, err := cmd.Flags().GetBool("quiet")
	if err != nil {
		return err
	}

	listArgs := []string{"ls"}
	if quiet {
		listArgs = append(listArgs, "-q")
	}

	_, stderr := common.Output(exec.Command(common.DM, listArgs...))
	if stderr != "" {
		return errors.New(stderr)
	}

	return nil
}

func runExample(cmd *cobra.Command, args []string) {
	const example = `
myctl:
  image: 10.10.1.12:5000/myctl:latest
  web: 10.10.1.12:5000/chiwen-web:master
  options:
  - "--combo=LITE"
  
inescure-resgitry:
  - 10.10.1.12:5000
  - 10.10.1.13:5000
  - 10.10.1.14:5000
  
master:
  name: devops160
  external_ip: 10.10.1.160
  internal_ip: 172.16.88.160
  os: ubuntu16.04
  docker: docker17.12.1
  
hosts:
- name: devops161
  external_ip: 10.10.1.161
  internal_ip: 172.16.88.161
  os: ubuntu16.04
  docker: docker17.12.1
- name: devops162
  external_ip: 10.10.1.162
  internal_ip: 172.16.88.162
  os: ubuntu16.04
  docker: docker17.12.1
- name: devops163
  external_ip: 10.10.1.163
  internal_ip: 172.16.88.163
  os: centos7
  docker: docker17.12.1
- name: devops164
  external_ip: 10.10.1.164
  internal_ip: 172.16.88.164
  os: centos7
  docker: docker17.12.1
  
clusters:
- name: red
  kind: swarm
  nodes:
  - name: devops161
	role: leader
  - name: devops162
	role: worker
- name: blue
  kind: kubernetes
  parameters:
	network: flannel
	elastic: on
  nodes:
  - name: devops164
	role: leader
  - name: devops165
	role: worker
`

	fmt.Println(example)
}

func runParse(cmd *cobra.Command, args []string) error {
	path, err := cmd.Flags().GetString("file")
	if err != nil {
		return err
	}

	deploy, err := common.ParseDeployment(path)
	if err != nil {
		return err
	}

	for _, c := range deploy.Clusters {
		for _, node := range c.Nodes {
			fmt.Printf("%+v\n", node)
		}
	}

	return nil
}
