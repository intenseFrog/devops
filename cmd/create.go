package cmd

import (
	"time"

	"github.com/spf13/cobra"

	"mydevops/pkg"
)

func init() {
	createCmd := &cobra.Command{
		Use:  "create",
		Long: "Create a number of hosts",
		RunE: runCreate,
	}
	createCmd.Flags().Bool("force", false, "recreate existing hosts")
	createCmd.Flags().StringP("file", "f", "", "Specify the file path")

	RootCmd.AddCommand(createCmd)
}

func runCreate(cmd *cobra.Command, args []string) error {
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

	fl := pkg.NewFileLock(path)
	if err := fl.TryLock(1 * time.Hour); err != nil {
		return err
	}
	defer fl.Unlock()

	if force, _ := cmd.Flags().GetBool("force"); force {
		deploy.Delete()
	}

	return deploy.Create()
}
