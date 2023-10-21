package cmd

import (
	"fmt"

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
		var (
			ff  string
			err error
		)
		if ff, err = cabriui.MutualExcludeFlags(
			[]string{"cnx", "ls", "purge", "clone", "asolf", "put", "get", "rename", "delete"},
			s3ToolsOptions.S3Session, s3ToolsOptions.S3List, s3ToolsOptions.S3Purge,
			s3ToolsOptions.S3Clone, s3ToolsOptions.S3AsOlf,
			s3ToolsOptions.S3Put, s3ToolsOptions.S3Get, s3ToolsOptions.S3Rename, s3ToolsOptions.S3Delete); err != nil {
			return err
		}
		if ff == "" {
			return fmt.Errorf("at least one operation option must be given with the s3Tools command")
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
	s3ToolsCmd.Flags().BoolVar(&s3ToolsOptions.S3AsOlf, "asolf", false, "clone swift container or s3 bucket to local olf")
	s3ToolsCmd.Flags().BoolVar(&s3ToolsOptions.S3Put, "put", false, "upload content as s3 object")
	s3ToolsCmd.Flags().BoolVar(&s3ToolsOptions.S3Get, "get", false, "download content from s3 object")
	s3ToolsCmd.Flags().BoolVar(&s3ToolsOptions.S3Rename, "rename", false, "rename s3 object")
	s3ToolsCmd.Flags().BoolVar(&s3ToolsOptions.S3Delete, "delete", false, "delete s3 object")
}
