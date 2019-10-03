package cmd

import (
	"errors"

	"github.com/spf13/cobra"

	"mydevops/pkg"
)

func init() {
	listCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List hosts",
		RunE:    runList,
	}
	listCmd.Flags().BoolP("quiet", "q", false, "List names only")

	RootCmd.AddCommand(listCmd)
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

	_, stderr := pkg.Output(pkg.Exec(pkg.DM, listArgs...))
	if stderr != "" {
		return errors.New(stderr)
	}

	return nil
}
