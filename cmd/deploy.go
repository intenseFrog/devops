package cmd

import (
	"time"

	"github.com/spf13/cobra"

	"mydevops/pkg"
)

func init() {
	deployCmd := &cobra.Command{
		Use:   "deploy",
		Short: "Deploy miaoyun",
		RunE:  runDeploy,
	}

	setFileFlags(deployCmd.Flags())
	RootCmd.AddCommand(deployCmd)
}

func runDeploy(cmd *cobra.Command, args []string) error {
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

	return deploy.Deploy()
}
