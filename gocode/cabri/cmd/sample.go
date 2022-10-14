package cmd

import (
	"fmt"

	"github.com/muesli/coral"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/cabriui"
)

var sampleOptions cabriui.SampleOptions

var sampleCmd = &coral.Command{
	Use:   "sample",
	Short: "just a sample",
	Long:  "just a sample",
	Args: func(cmd *coral.Command, args []string) error {
		returnUsageAndErr := func(err error) error {
			cmd.UsageFunc()(cmd)
			return err
		}
		if len(args) == 0 {
			return returnUsageAndErr(fmt.Errorf("at least one arg please"))
		}
		return nil
	},
	Hidden:       true,
	SilenceUsage: true,
	RunE: func(cmd *coral.Command, args []string) error {
		sampleOptions.BaseOptions = baseOptions
		return cabriui.CLIRun[cabriui.SampleOptions, cabriui.SampleVars](
			cmd.InOrStdin(), cmd.OutOrStdout(), cmd.ErrOrStderr(),
			sampleOptions, args,
			cabriui.SampleStartup, cabriui.SampleShutdown)
	},
}

func init() {
	cliCmd.AddCommand(sampleCmd)
	sampleCmd.Flags().BoolVarP(&sampleOptions.FlagSample, "fs", "f", false, "just a sample for a flag")
}
