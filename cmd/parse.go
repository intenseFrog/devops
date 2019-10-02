package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"mydevops/pkg"
)

func init() {
	parseCmd := &cobra.Command{
		Use:    "parse",
		Hidden: true,
		RunE:   runParse,
	}
	parseCmd.Flags().StringP("file", "f", "", "Specify the file path")

	RootCmd.AddCommand(parseCmd)
}

func runParse(cmd *cobra.Command, args []string) error {
	path, err := cmd.Flags().GetString("file")
	if err != nil {
		return err
	}

	deploy, err := pkg.ParseDeployment(path)
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
