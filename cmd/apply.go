package cmd

import (
	"time"

	"github.com/spf13/cobra"

	"mydevops/pkg"
)

func init() {
	applyCmd := &cobra.Command{
		Use:   "apply",
		Short: "create machines and deploy miaoyun",
		Long:  "create machines and deploy miaoyun",
		RunE:  runApply,
	}
	applyCmd.Flags().Bool("force", false, "destroy previous machines")
	applyCmd.Flags().StringP("file", "f", "", "Specify the file path")

	RootCmd.AddCommand(applyCmd)
}

func runApply(cmd *cobra.Command, args []string) error {
	start := time.Now()
	defer pkg.PrintDone(start)

	path, err := cmd.Flags().GetString("file")
	if err != nil {
		return err
	}

	deploy, err := pkg.ParseDeployment(path)
	if err != nil {
		return err
	}

	if force, _ := cmd.Flags().GetBool("force"); force {
		deploy.Delete()
	}

	lock, err := pkg.FileLock(path)
	if err != nil {
		return err
	}
	defer lock.Unlock()

	if err = deploy.Create(); err != nil {
		return err
	}

	if err = deploy.Deploy(); err != nil {
		return err
	}

	return nil
}
