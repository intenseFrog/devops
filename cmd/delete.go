package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"mydevops/pkg"
)

func init() {
	deleteCmd := &cobra.Command{
		Use:     "delete",
		Aliases: []string{"destroy", "remove", "rm"},
		Short:   "delete nodes defined by yaml file",
		Long:    "delete nodes defined by yaml file",
		RunE:    runDelete,
	}
	deleteCmd.Flags().StringP("file", "f", "", "Specify the file path")
	deleteCmd.Flags().Bool("force", false, "force deleting machines")

	RootCmd.AddCommand(deleteCmd)
}

func runDelete(cmd *cobra.Command, args []string) error {
	path, err := cmd.Flags().GetString("file")
	if err != nil {
		return err
	}

	deploy, err := pkg.ParseDeployment(path)
	if err != nil {
		return err
	}

	var names []string
	for _, h := range deploy.ListHosts() {
		names = append(names, h.Name)
	}

	force, _ := cmd.Flags().GetBool("force")
	msg := fmt.Sprintf("About to remove %s", strings.Join(names, ", "))

	lock, err := pkg.FileLock(path)
	if err != nil {
		return err
	}
	defer lock.Unlock()

	if force || pkg.Confirm(msg) {
		deploy.Delete()
	}

	return nil
}
