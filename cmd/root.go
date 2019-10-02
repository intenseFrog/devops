package cmd

import "github.com/spf13/cobra"

var (
	RootCmd = &cobra.Command{
		Use:     "mydevops",
		Aliases: []string{"myd"},
		Long:    "a command line tool to enable the CI/CD of Miaoyun",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
)
