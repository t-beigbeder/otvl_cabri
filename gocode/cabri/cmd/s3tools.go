package cmd

import (
	"github.com/muesli/coral"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/cabriui"
)

var s3ToolsOptions cabriui.S3ToolsOptions

var s3ToolsCmd = &coral.Command{
	Use:   "s3tools",
	Short: "check various configuration",
	Long:  "check various configuration",
	Args: func(cmd *coral.Command, args []string) error {
		return nil
	},
	RunE: func(cmd *coral.Command, args []string) error {
		s3ToolsOptions.BaseOptions = baseOptions
		if err := cabriui.MutualExcludeFlags(
			[]string{"cnx", "ls", "purge", "clone"},
			s3ToolsOptions.S3Session, s3ToolsOptions.S3List, s3ToolsOptions.S3Purge, s3ToolsOptions.S3Clone); err != nil {
			return err
		}
		return cabriui.CLIRun[cabriui.S3ToolsOptions, *cabriui.S3ToolsVars](
			cmd.InOrStdin(), cmd.OutOrStdout(), cmd.ErrOrStderr(),
			s3ToolsOptions, args,
			cabriui.S3ToolsStartup, cabriui.S3ToolsShutdown)
	},
	SilenceUsage: true,
}

func init() {
	cliCmd.AddCommand(s3ToolsCmd)
	s3ToolsCmd.Flags().BoolVar(&s3ToolsOptions.S3Session, "cnx", false, "checks connexion with given S3 parameters")
	s3ToolsCmd.Flags().BoolVar(&s3ToolsOptions.S3List, "ls", false, "list S3 objects with given prefix")
	s3ToolsCmd.Flags().BoolVar(&s3ToolsOptions.S3Purge, "purge", false, "purge swift container or s3 bucket")
	s3ToolsCmd.Flags().BoolVar(&s3ToolsOptions.S3Clone, "clone", false, "clone swift container or s3 bucket to another")
}
