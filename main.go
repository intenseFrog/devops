package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"mydevops/common"

	"github.com/spf13/cobra"
)

// CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o mydevops

func main() {
	RootCmd := &cobra.Command{
		Use:   "mydevops",
		Short: "CLI tool to manage miaoyun",
		Long:  "CLI tool to manage miaoyun",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	cleanKnowHosts := &cobra.Command{
		Use:   "clean-known",
		Short: "clean .ssh/known_hosts",
		Long:  "clean .ssh/known_hosts",
		RunE:  runCleanKnowHosts,
	}
	cleanKnowHosts.Flags().StringP("file", "f", "", "Specify the file path")

	createCmd := &cobra.Command{
		Use:   "create",
		Short: "create a bunch of machines",
		Long:  "create a bunch of machines",
		RunE:  runCreate,
	}
	createCmd.Flags().StringP("file", "f", "", "Specify the file path")
	createCmd.MarkFlagRequired("file")

	exampleCmd := &cobra.Command{
		Use:   "example",
		Short: "print out an example of a yaml file",
		Long:  "print out an example of a yaml file",
		Run:   runExample,
	}

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
		Short: "destroy one or more nodes",
		Long:  "destroy one or more nodes",
		RunE:  runDestroy,
	}
	destroyCmd.Flags().StringP("file", "f", "", "Specify the file path")
	destroyCmd.Flags().BoolP("yes", "y", false, "Assume automatic yes on removing machines")
	destroyCmd.Flags().Bool("all", false, "Remove all the machines available")
	// destroyCmd.Flags().Boo("file", "f", "", "Specify the file path")

	licenseCmd := &cobra.Command{
		Use:   "license",
		Short: "license a node",
		RunE:  runLicense,
	}
	licenseCmd.Flags().StringP("file", "f", "", "Specify the file path")

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

	RootCmd.AddCommand(cleanKnowHosts)
	RootCmd.AddCommand(createCmd)
	RootCmd.AddCommand(deployCmd)
	RootCmd.AddCommand(exampleCmd)
	RootCmd.AddCommand(licenseCmd)
	RootCmd.AddCommand(listCmd)
	RootCmd.AddCommand(destroyCmd)
	RootCmd.AddCommand(updateCmd)

	// back door
	RootCmd.AddCommand(parseCmd)

	if err := RootCmd.Execute(); err != nil {
		os.Exit(-1)
	}
}

func runCleanKnowHosts(cmd *cobra.Command, args []string) error {
	path, err := cmd.Flags().GetString("file")
	if err != nil {
		return err
	}

	deployment, err := common.ParseDeployment(path)
	if err != nil {
		return err
	}

	for _, c := range deployment.Clusters {
		c.CleanKnownHosts()
	}

	return nil
}

func runCreate(cmd *cobra.Command, args []string) error {
	start := time.Now()

	path, err := cmd.Flags().GetString("file")
	if err != nil {
		return err
	}

	deployment, err := common.ParseDeployment(path)
	if err != nil {
		return err
	}

	if err := deployment.Create(); err != nil {
		return err
	}

	common.PrintDone(start)
	return nil
}

func runDeploy(cmd *cobra.Command, args []string) error {
	start := time.Now()

	path, err := cmd.Flags().GetString("file")
	if err != nil {
		return err
	}

	deployment, err := common.ParseDeployment(path)
	if err != nil {
		return err
	}

	var names []string
	for _, n := range deployment.ListNodes() {
		names = append(names, n.Name)
	}
	common.Destroy(names, true)

	if err := deployment.Create(); err != nil {
		return err
	}

	common.RemoveKnownHosts()

	if err := deployment.Deploy(); err != nil {
		return err
	}

	common.PrintDone(start)
	return nil
}

func runDestroy(cmd *cobra.Command, args []string) error {
	var names []string

	all, err := cmd.Flags().GetBool("all")
	if err != nil {
		return err
	}

	path, err := cmd.Flags().GetString("file")
	if err != nil {
		return err
	}

	if all && path != "" {
		return errors.New("Cannot specify --all and --file at same time")
	}

	if path != "" {
		if deployment, err := common.ParseDeployment(path); err == nil {
			for _, n := range deployment.ListNodes() {
				names = append(names, n.Name)
			}
		} else {
			return err
		}
	} else if all {
		output, stderr := common.Output(exec.Command(common.DM, "ls", "-q"))
		if stderr != "" {
			return errors.New(stderr)
		}
		names = strings.Split(output, "\n")
	} else {
		names = args
	}

	yes, err := cmd.Flags().GetBool("yes")
	if err != nil {
		return err
	}

	common.Destroy(names, yes)
	return nil
}

func runUpdate(cmd *cobra.Command, args []string) error {
	start := time.Now()

	path, err := cmd.Flags().GetString("file")
	if err != nil {
		return err
	}

	deployment, err := common.ParseDeployment(path)
	if err != nil {
		return err
	}

	if err := deployment.Update(); err != nil {
		return err
	}

	common.PrintDone(start)
	return nil
}

func runLicense(cmd *cobra.Command, args []string) error {
	path, err := cmd.Flags().GetString("file")
	if err != nil {
		return err
	}

	deploy, err := common.ParseDeployment(path)
	if err == nil {
		return err
	}

	return deploy.License()
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

	output, stderr := common.Output(exec.Command(common.DM, listArgs...))
	if stderr != "" {
		return errors.New(stderr)
	}

	fmt.Println(output)
	return nil
}

func runExample(cmd *cobra.Command, args []string) {
	const example = `
myctl:
  image: 10.10.1.12:5000/myctl:latest
  channel: devops
  
clusters:
- name: default
  nodes:
  - name: devops160
	role: master
	external_ip: 10.10.1.160
	internal_ip: 172.16.88.160
	os: ubuntu16.04
	docker: docker17.12.1
- name: red
  kind: swarm
  nodes:
  - name: devops161
	role: leader
	external_ip: 10.10.1.161
	internal_ip: 172.16.88.161
	os: ubuntu16.04
	docker: docker17.12.1
  - name: devops162
	role: worker
	external_ip: 10.10.1.162
	internal_ip: 172.16.88.162
	os: ubuntu16.04
	docker: docker17.12.1
- name: blue
  kind: kubernetes
  parameters:
	network: flannel
	elastic: on
  nodes:
  - name: devops164
	role: leader
	external_ip: 10.10.1.164
	internal_ip: 172.16.88.164
	os: centos7
	docker: docker17.12.1
  - name: devops165
	role: worker
	external_ip: 10.10.1.165
	internal_ip: 172.16.88.165
	os: ubuntu16.04
	docker: docker17.12.1
`

	fmt.Println(example)
}

func runParse(cmd *cobra.Command, args []string) error {
	path, err := cmd.Flags().GetString("file")
	if err != nil {
		return err
	}

	deployment, err := common.ParseDeployment(path)
	if err != nil {
		return err
	}

	for i := range deployment.Clusters {
		cluster := deployment.Clusters[i]
		cluster.Normalize()
	}

	for _, c := range deployment.Clusters {
		for _, node := range c.Nodes {
			fmt.Println(node.String())
		}
	}

	return nil
}
