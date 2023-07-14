package cmd

import (
	"fmt"

	"github.com/muesli/coral"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/cabriui"
)

var scheduleOptions cabriui.ScheduleOptions

var scheduleCmd = &coral.Command{
	Use:   "schedule",
	Short: "schedule periodic and run http triggered actions",
	Long:  "schedule periodic and run http triggered actions",
	RunE: func(cmd *coral.Command, args []string) error {
		if scheduleOptions.SpecFile == "" {
			cmd.UsageFunc()(cmd)
			return fmt.Errorf("a specification file must be provided")
		}
		scheduleOptions.BaseOptions = baseOptions
		return cabriui.CLIRun[cabriui.ScheduleOptions, *cabriui.ScheduleVars](
			cmd.InOrStdin(), cmd.OutOrStdout(), cmd.ErrOrStderr(),
			scheduleOptions, args,
			cabriui.ScheduleStartup, cabriui.ScheduleShutdown)
	},
	SilenceUsage: true,
}

func init() {
	rootCmd.AddCommand(scheduleCmd)
	scheduleCmd.Flags().StringVar(&baseOptions.ConfigDir, "cdir", "", "load configuration files from this directory instead of .cabri in home directory")
	scheduleCmd.Flags().StringVarP(&baseOptions.PassFile, "pfile", "", "", "file containing the master password")
	scheduleCmd.Flags().BoolVar(&baseOptions.Password, "password", false, "force master password prompt")
	scheduleCmd.Flags().BoolVar(&scheduleOptions.HasLog, "haslog", false, "output http access log for the API")
	scheduleCmd.Flags().StringVarP(&scheduleOptions.SpecFile, "sfile", "s", "", "file containing the scheduling specification")
	scheduleCmd.Flags().BoolVar(&scheduleOptions.HasHttp, "http", false, "launches an http server to trigger updates or report status")
	scheduleCmd.Flags().StringVarP(&scheduleOptions.Address, "address", "", ":3000", "host:port to listen to, defaults :3000")
}
