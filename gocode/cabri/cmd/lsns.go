package cmd

import (
	"fmt"

	"github.com/muesli/coral"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/cabriui"
)

var lsnsOptions cabriui.LsnsOptions

var lsnsCmd = &coral.Command{
	Use:   "lsns <dss-type:/path/to/dss@path/in/dss>",
	Short: "list namespace information",
	Long:  `list namespace information`,
	Args: func(cmd *coral.Command, args []string) error {
		if len(args) != 1 {
			cmd.UsageFunc()(cmd)
			return fmt.Errorf("a DSS namespace must be provided")
		}
		_, _, _, err := cabriui.CheckDssPath(args[0])

		if err != nil {
			cmd.UsageFunc()(cmd)
			return fmt.Errorf("%v\nsyntax: dss-type:/path/to/dss@path/in/dss\nfor instance\n\tfsy:/home/guest@Downloads", err)
		}
		return nil
	},
	RunE: func(cmd *coral.Command, args []string) error {
		lsnsOptions.BaseOptions = baseOptions
		if _, err := cabriui.CheckTimeStamp(lsnsOptions.LastTime); err != nil {
			return err
		}
		dssType, _, _, _ := cabriui.CheckDssPath(args[0])
		if dssType == "fsy" && lsnsOptions.Checksum {
			lsnsOptions.RedLimit = 0
		}
		return cabriui.CLIRun[cabriui.LsnsOptions, *cabriui.LsnsVars](
			cmd.InOrStdin(), cmd.OutOrStdout(), cmd.ErrOrStderr(),
			lsnsOptions, args,
			cabriui.LsnsStartup, cabriui.LsnsShutdown)
	},
	SilenceUsage: true,
}

func init() {
	cliCmd.AddCommand(lsnsCmd)
	lsnsCmd.Flags().BoolVarP(&lsnsOptions.Recursive, "recursive", "r", false, "recursively list subnamespaces information")
	lsnsCmd.Flags().BoolVarP(&lsnsOptions.Sorted, "sorted", "s", false, "sort entries by name")
	lsnsCmd.Flags().BoolVarP(&lsnsOptions.Time, "time", "t", false, "sort entries by last modification date and time")
	lsnsCmd.Flags().BoolVarP(&lsnsOptions.Long, "long", "l", false, "long format display")
	lsnsCmd.Flags().BoolVarP(&lsnsOptions.Checksum, "checksum", "c", false, "calculate content's checksum if not available and display it")
	lsnsCmd.Flags().BoolVar(&lsnsOptions.Reverse, "reverse", false, "sort is reversed")
	lsnsCmd.Flags().StringVar(&lsnsOptions.LastTime, "lasttime", "", "upper time of entries retrieved in historized DSS")
}
