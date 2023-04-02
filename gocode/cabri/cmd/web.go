package cmd

import (
	"fmt"

	"github.com/muesli/coral"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/cabriui"
)

var webApiOptions cabriui.WebApiOptions

var webApiCmd = &coral.Command{
	Use:   "webapi",
	Short: "launches web API servers",
	Long:  `launches web API servers dealing with local files or cloud object storage data`,
	Args: func(cmd *coral.Command, args []string) error {
		returnUsageAndErr := func(err error) error {
			cmd.UsageFunc()(cmd)
			return err
		}
		checkArg := func(arg string) error {
			if _, _, _, _, _, err := cabriui.CheckDssUrlMapping(arg); err != nil {
				return fmt.Errorf(`
%v
DSS URL mapping syntax is: dss-type+http[s]://server:port/local/path@root
for instance
	obs+http://localhost:3000/data/local/obs1@obs1`,
					err,
				)
			}
			return nil
		}
		if len(args) == 0 {
			return returnUsageAndErr(fmt.Errorf("no DSS to URL mapping was provided as command argument"))
		}
		for i := 0; i < len(args); i++ {
			if err := checkArg(args[i]); err != nil {
				return returnUsageAndErr(err)
			}
		}
		return nil
	},
	RunE: func(cmd *coral.Command, args []string) error {
		webApiOptions.BaseOptions = baseOptions
		return cabriui.CLIRun[cabriui.WebApiOptions, *cabriui.WebApiVars](
			cmd.InOrStdin(), cmd.OutOrStdout(), cmd.ErrOrStderr(),
			webApiOptions, args,
			cabriui.WebApiStartup, cabriui.WebApiShutdown)
	},
	SilenceUsage: true,
}

func init() {
	rootCmd.AddCommand(webApiCmd)
	webApiCmd.Flags().StringVar(&baseOptions.ConfigDir, "cdir", "", "load configuration files from this directory instead of .cabri in home directory")
	webApiCmd.Flags().StringVarP(&baseOptions.PassFile, "pfile", "", "", "file containing the master password")
	webApiCmd.Flags().BoolVar(&baseOptions.Password, "password", false, "force master password prompt")
	webApiCmd.Flags().StringArrayVar(&baseOptions.IndexImplems, "ximpl", nil, "list of non-default object storage index implementations")
	webApiCmd.Flags().StringArrayVar(&baseOptions.ObsRegions, "obsrg", nil, "list of object storage regions")
	webApiCmd.Flags().StringArrayVar(&baseOptions.ObsEndpoints, "obsep", nil, "list of object storage endpoints")
	webApiCmd.Flags().StringArrayVar(&baseOptions.ObsContainers, "obsct", nil, "list of object storage containers")
	webApiCmd.Flags().StringArrayVar(&baseOptions.ObsAccessKeys, "obsak", nil, "list of object storage access keys")
	webApiCmd.Flags().StringArrayVar(&baseOptions.ObsSecretKeys, "obssk", nil, "list of object storage secret keys")
	webApiCmd.Flags().StringVar(&baseOptions.TlsCert, "tlscrt", "", "certificate file on https server or untrusted CA on https client")
	webApiCmd.Flags().BoolVar(&baseOptions.TlsNoCheck, "tlsnc", false, "no check of certificate by https client")
	webApiCmd.Flags().BoolVar(&webApiOptions.HasLog, "haslog", false, "output http access log for the API")
	webApiCmd.Flags().StringVar(&webApiOptions.TlsKey, "tlskey", "", "certificate key file")
}
