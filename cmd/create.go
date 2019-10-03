package cmd

import (
	"time"

	"github.com/spf13/cobra"

	"mydevops/pkg"
)

func init() {
	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create hosts",
		RunE:  runCreate,
	}

	setFileFlags(createCmd.Flags())
	RootCmd.AddCommand(createCmd)
}

func runCreate(cmd *cobra.Command, args []string) error {
	start := time.Now()
	defer pkg.PrintDone(start)

	flags := cmd.Flags()
	path, err := flags.GetString("file")
	if err != nil {
		return err
	}

	deploy, err := pkg.ParseDeployment(path)
	if err != nil {
		return err
	}

	if lock, _ := flags.GetBool("lock"); lock {
		timeout, err := flags.GetString("lock-timeout")
		if err != nil {
			return err
		}

		d, err := parseDuration(timeout)
		if err != nil {
			return err
		}

		fl := pkg.NewFileLock(path, d)
		if err := fl.Lock(); err != nil {
			return err
		}

		defer fl.Unlock()
	}

	if force, _ := flags.GetBool("force"); force {
		deploy.Delete()
	}

	return deploy.Create()
}
