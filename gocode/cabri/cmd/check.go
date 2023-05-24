package cmd

import (
	"github.com/muesli/coral"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/cabriui"
)

var checkOptions cabriui.CheckOptions

var checkCmd = &coral.Command{
	Use:   "check",
	Short: "check various configuration",
	Long:  "check various configuration",
	Args: func(cmd *coral.Command, args []string) error {
		return nil
	},
	RunE: func(cmd *coral.Command, args []string) error {
		checkOptions.BaseOptions = baseOptions
		if err := cabriui.MutualExcludeFlags(
			[]string{"s3cnx", "s3ls"},
			checkOptions.S3Session, checkOptions.S3List); err != nil {
			return err
		}
		return cabriui.CLIRun[cabriui.CheckOptions, *cabriui.CheckVars](
			cmd.InOrStdin(), cmd.OutOrStdout(), cmd.ErrOrStderr(),
			checkOptions, args,
			cabriui.CheckStartup, cabriui.CheckShutdown)
	},
	SilenceUsage: true,
}

func init() {
	cliCmd.AddCommand(checkCmd)
	checkCmd.Flags().BoolVar(&checkOptions.S3Session, "s3cnx", false, "checks connexion with given S3 parameters")
	checkCmd.Flags().BoolVar(&checkOptions.S3List, "s3ls", false, "list S3 objects with given prefix")
}
