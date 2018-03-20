package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

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

	deployCmd := &cobra.Command{
		Use:   "deploy",
		Short: "create and deploy a cluster",
		Long:  "create and deploy a cluster",
		RunE:  runDeploy,
	}
	deployCmd.Flags().StringP("file", "f", "", "Specify the file path")

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

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "list nodes",
		Long:  "list nodes",
		RunE:  runList,
	}
	listCmd.Flags().BoolP("quiet", "q", false, "List names only")

	RootCmd.AddCommand(cleanKnowHosts)
	RootCmd.AddCommand(deployCmd)
	RootCmd.AddCommand(listCmd)
	RootCmd.AddCommand(destroyCmd)

	if err := RootCmd.Execute(); err != nil {
		os.Exit(-1)
	}
}

func runDeploy(cmd *cobra.Command, args []string) error {
	path, err := cmd.Flags().GetString("file")
	if err != nil {
		return err
	}

	deployment := &common.Deployment{}
	if err := deployment.Parse(path); err != nil {
		return err
	}

	for _, node := range deployment.Nodes {
		if err := node.Create(); err != nil {
			return err
		}
	}

	defer common.Elite("logout")
	// time.Sleep(30 * time.Second)

	for _, node := range deployment.Nodes {
		var err error
		switch role := node.Role; role {
		case "master":
			fmt.Println("Licensing....")
			if err = node.License(); err == nil {
				fmt.Println("Deploying....")
				err = node.Deploy()
			}
		case "leader":
			fmt.Println("Initializing....")
			if clusterNode, err := node.ClusterNode(); err == nil {
				err = clusterNode.Init()
			}
		case "worker":
			fmt.Println("Joining....")
			if clusterNode, err := node.ClusterNode(); err == nil {
				err = clusterNode.Join()
			}
		default:
			err = fmt.Errorf("unknown role: %s", role)
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func runList(cmd *cobra.Command, args []string) error {
	quiet, err := cmd.Flags().GetBool("quiet")
	if err != nil {
		return err
	}

	arguments := []string{"list", "--all"}
	if quiet {
		arguments = append(arguments, "--name")
	}

	output, stderr := common.Output(exec.Command("virsh", arguments...))
	if stderr != "" {
		return errors.New(stderr)
	}

	fmt.Println(output)
	return nil
}

func runCreate(cmd *cobra.Command, args []string) error {
	path, err := cmd.Flags().GetString("file")
	if err != nil {
		return err
	}

	deployment := &common.Deployment{}
	if err := deployment.Parse(path); err != nil {
		return err
	}

	for _, node := range deployment.Nodes {
		if err := node.Create(); err != nil {
			return err
		}
	}

	return nil
}

func runCleanKnowHosts(cmd *cobra.Command, args []string) error {
	path, err := cmd.Flags().GetString("file")
	if err != nil {
		return err
	}

	deployment := &common.Deployment{}
	if err := deployment.Parse(path); err != nil {
		return err
	}

	for _, node := range deployment.Nodes {
		node.CleanKnownHost()
	}

	return nil
}

func runDestroy(cmd *cobra.Command, args []string) error {
	var nodes []*common.Node
	names := make([]string, 0)

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

	if all {
		output, stderr := common.Output(exec.Command("virsh", "list", "--all", "--name"))
		if stderr != "" {
			return errors.New(stderr)
		}
		names = strings.Split(output, "\n")
		for _, name := range names {
			nodes = append(nodes, &common.Node{Name: name})
		}
	} else if path == "" {
		names = args
		for _, name := range names {
			nodes = append(nodes, &common.Node{Name: name})
		}
	} else {
		deployment := &common.Deployment{}
		if err := deployment.Parse(path); err != nil {
			return err
		}
		nodes = deployment.Nodes
		for _, n := range nodes {
			names = append(names, n.Name)
		}
	}

	yes, err := cmd.Flags().GetBool("yes")
	if err != nil {
		return err
	}

	msg := fmt.Sprintf("About to remove %s", strings.Join(names, ", "))
	if yes || common.Confirm(msg) {
		for _, node := range nodes {
			if err := node.Destroy(); err != nil {
				fmt.Printf("%s removal failed: %s\n", node.Name, err.Error())
			}
		}
	}

	return nil
}
