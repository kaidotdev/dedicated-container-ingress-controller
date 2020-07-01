package cmd

import "github.com/spf13/cobra"

func GetRootCmd(args []string) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "dedicated-container-ingress-controller",
		Short:        "",
		SilenceUsage: true,
	}

	cmd.SetArgs(args)
	cmd.AddCommand(serverCmd())

	return cmd
}
