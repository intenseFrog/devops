package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

// CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o mydevops

var pathFile *string

func main() {
	RootCmd := &cobra.Command{
		Use:   "mydevops",
		Short: "CLI tool to manage miaoyun",
		Long:  "CLI tool to manage miaoyun",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	createCmd := &cobra.Command{
		Use:   "create",
		Short: "create a cluster",
		Long:  "create a cluster",
		RunE:  runCreate,
	}
	pathFile = createCmd.Flags().StringP("file", "f", "", "Specify the file path")

	destroyCmd := &cobra.Command{
		Use:   "destroy [nodes]",
		Short: "destroy one or more nodes",
		Long:  "destroy one or more nodes",
		RunE:  runDestroy,
	}

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "list nodes",
		Long:  "list nodes",
		RunE:  runList,
	}

	RootCmd.AddCommand(createCmd)
	RootCmd.AddCommand(listCmd)
	RootCmd.AddCommand(destroyCmd)

	if err := RootCmd.Execute(); err != nil {
		os.Exit(-1)
	}
}

func runCreate(cmd *cobra.Command, args []string) error {
	deployment, err := Parse(*pathFile)
	if err != nil {
		return err
	}

	return deployment.Create()
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
	return nil
}
