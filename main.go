package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"time"

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

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "list nodes",
		Long:  "list nodes",
		RunE:  runList,
	}

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

	deployment, err := Parse(path)
	if err != nil {
		return err
	}

	for _, node := range deployment.Nodes {
		if err := node.Create(); err != nil {
			return err
		}
	}

	time.Sleep(30 * time.Second)

	for _, node := range deployment.Nodes {

		node.CleanKnownHost()

		var err error
		switch role := node.Role; role {
		case "master":
			fmt.Println("Licensing....")
			if err = node.License(); err == nil {
				fmt.Println("Deploying....")
				err = node.Deploy(deployment.Myctl)
			}
		case "leader":
			fmt.Println("Initializing....")
			err = node.Init()
		case "worker":
			err = node.Join()
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
	output, stderr := Output(exec.Command("virsh", "list", "--all"))
	if stderr != "" {
		return errors.New(stderr)
	}

	fmt.Println(output)
	return nil
}

func runDestroy(cmd *cobra.Command, args []string) error {
	var nodes []*Node
	path, err := cmd.Flags().GetString("file")
	if err != nil {
		return err
	}

	if path == "" {
		for _, name := range args {
			nodes = append(nodes, &Node{Name: name})
		}
	} else {
		deployment, err := Parse(path)
		if err != nil {
			return err
		}
		nodes = deployment.Nodes
	}

	for _, node := range nodes {
		if err = node.Destroy(); err != nil {
			fmt.Printf("%s removal failed: %s\n", node.Name, err.Error())
		}
	}

	return err
}
