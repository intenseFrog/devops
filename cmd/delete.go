package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"mydevops/pkg"
)

func init() {
	deleteCmd := &cobra.Command{
		Use:     "delete",
		Aliases: []string{"destroy", "remove", "rm"},
		Short:   "Delete hosts",
		RunE:    runDelete,
	}

	setFileFlags(deleteCmd.Flags())
	RootCmd.AddCommand(deleteCmd)
}

func runDelete(cmd *cobra.Command, args []string) error {
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

	var names []string
	for _, h := range deploy.ListHosts() {
		names = append(names, h.Name)
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

	force, _ := flags.GetBool("force")
	msg := fmt.Sprintf("About to remove %s", strings.Join(names, ", "))
	if force || pkg.Confirm(msg) {
		deploy.Delete()
	}

	return nil
}
