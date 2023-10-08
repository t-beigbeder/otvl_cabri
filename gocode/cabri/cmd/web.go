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

var restApiCmd = &coral.Command{
	Use:   "rest",
	Short: "launches REST API server",
	Long:  `launches REST API server dealing with local files or cloud object storage data`,
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
		webApiOptions.IsRest = true
		return cabriui.CLIRun[cabriui.WebApiOptions, *cabriui.WebApiVars](
			cmd.InOrStdin(), cmd.OutOrStdout(), cmd.ErrOrStderr(),
			webApiOptions, args,
			cabriui.WebApiStartup, cabriui.WebApiShutdown)
	},
	SilenceUsage: true,
}

func init() {
	rootCmd.AddCommand(webApiCmd)
	webApiCmd.PersistentFlags().StringVar(&baseOptions.ConfigDir, "cdir", "", "load configuration files from this directory instead of .cabri in home directory")
	webApiCmd.PersistentFlags().StringVarP(&baseOptions.PassFile, "pfile", "", "", "file containing the master password")
	webApiCmd.PersistentFlags().BoolVar(&baseOptions.Password, "password", false, "force master password prompt")
	webApiCmd.PersistentFlags().StringArrayVar(&baseOptions.IndexImplems, "ximpl", nil, "list of non-default object storage index implementations")
	webApiCmd.PersistentFlags().StringArrayVar(&baseOptions.ObsRegions, "obsrg", nil, "list of object storage regions")
	webApiCmd.PersistentFlags().StringArrayVar(&baseOptions.ObsEndpoints, "obsep", nil, "list of object storage endpoints")
	webApiCmd.PersistentFlags().StringArrayVar(&baseOptions.ObsContainers, "obsct", nil, "list of object storage containers")
	webApiCmd.PersistentFlags().StringArrayVar(&baseOptions.ObsAccessKeys, "obsak", nil, "list of object storage access keys")
	webApiCmd.PersistentFlags().StringArrayVar(&baseOptions.ObsSecretKeys, "obssk", nil, "list of object storage secret keys")
	webApiCmd.PersistentFlags().StringVar(&baseOptions.TlsCert, "tlscrt", "", "certificate file on https server or untrusted CA on https client")
	webApiCmd.PersistentFlags().BoolVar(&baseOptions.TlsNoCheck, "tlsnc", false, "no check of certificate by https client")
	webApiCmd.PersistentFlags().BoolVar(&webApiOptions.HasLog, "haslog", false, "output http access log for the API")
	webApiCmd.PersistentFlags().StringVar(&webApiOptions.TlsKey, "tlskey", "", "certificate key file")
	webApiCmd.AddCommand(restApiCmd)
	restApiCmd.Flags().StringArrayVarP(&baseOptions.Users, "user", "u", nil, "list of ACL users for retrieval")
	restApiCmd.Flags().StringArrayVar(&baseOptions.ACL, "acl", nil, "list of ACL <user:rights> items (defaults to rw) for creation and update")
	restApiCmd.PersistentFlags().StringVar(&baseOptions.HUser, "huser", "", "http client user")
	restApiCmd.PersistentFlags().StringVar(&baseOptions.HPFile, "hpfile", "", "file containing the http client user password")
	restApiCmd.PersistentFlags().BoolVar(&baseOptions.HPassword, "hpassword", false, "force http client user password prompt")
	restApiCmd.Flags().StringVar(&webApiOptions.LastTime, "lasttime", "", "upper time of entries retrieved in historized DSS")
	restApiCmd.Flags().StringVar(&webApiOptions.TlsClientCert, "tlsclientcrt", "", "untrusted CA on https client")
}
