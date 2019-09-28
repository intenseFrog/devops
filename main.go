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

	createCmd := &cobra.Command{
		Use:   "create",
		Short: "create a bunch of machines",
		Long:  "create a bunch of machines",
		RunE:  runCreate,
	}
	createCmd.Flags().Bool("force", false, "destroy previous machines")
	createCmd.Flags().StringP("file", "f", "", "Specify the file path")

	deployCmd := &cobra.Command{
		Use:   "deploy",
		Short: "create and deploy a cluster",
		Long:  "create and deploy a cluster",
		RunE:  runDeploy,
	}
	deployCmd.Flags().StringP("file", "f", "", "Specify the file path")

	deleteCmd := &cobra.Command{
		Use:   "delete",
		Short: "delete nodes defined by yaml file",
		Long:  "delete nodes defined by yaml file",
		RunE:  runDelete,
	}
	deleteCmd.Flags().StringP("file", "f", "", "Specify the file path")
	deleteCmd.Flags().Bool("force", false, "force deleting machines")

	// deprecated, use delete instead
	destroyCmd := &cobra.Command{
		Use:    "destroy",
		Short:  "destroy nodes defined by yaml file (deprecated)",
		Long:   "destroy nodes defined by yaml file (deprecated)",
		Hidden: true,
		RunE:   runDelete,
	}
	destroyCmd.Flags().StringP("file", "f", "", "Specify the file path")
	destroyCmd.Flags().Bool("force", false, "force destroying machines")

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

	RootCmd.AddCommand(applyCmd)
	RootCmd.AddCommand(createCmd)
	RootCmd.AddCommand(deployCmd)
	RootCmd.AddCommand(listCmd)
	RootCmd.AddCommand(deleteCmd)
	RootCmd.AddCommand(destroyCmd)
	RootCmd.AddCommand(updateCmd)
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

func runDelete(cmd *cobra.Command, args []string) error {
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
