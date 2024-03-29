package cmd

import (
	"github.com/muesli/coral"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/cabriui"
)

var baseOptions cabriui.BaseOptions

var cliCmd = &coral.Command{
	Use:   "cli [subcommand]",
	Short: "Cabri command line interface",
	Long:  `Cabri command line interface calling subcommands`,
}

var dssCmd = &coral.Command{
	Use:   "dss [subcommand]",
	Short: "Cabri DSS management",
	Long:  "Cabri DSS management calling subcommands",
}

func init() {
	rootCmd.AddCommand(cliCmd)
	cliCmd.PersistentFlags().StringVar(&baseOptions.ConfigDir, "cdir", "", "load configuration files from this directory instead of .cabri in home directory")
	cliCmd.PersistentFlags().StringArrayVarP(&baseOptions.Users, "user", "u", nil, "list of ACL users for retrieval")
	cliCmd.PersistentFlags().StringArrayVar(&baseOptions.ACL, "acl", nil, "list of ACL <user:rights> items (defaults to rw) for creation and update")
	cliCmd.PersistentFlags().StringVar(&baseOptions.PassFile, "pfile", "", "file containing the master password")
	cliCmd.PersistentFlags().BoolVar(&baseOptions.Password, "password", false, "force master password prompt")
	cliCmd.PersistentFlags().BoolVar(&baseOptions.Serial, "serial", false, "run all tasks in sequence")
	cliCmd.PersistentFlags().IntVar(&baseOptions.MaxThread, "maxt", 0, "sets the maximum OS thread number, defaults to 10000")
	cliCmd.PersistentFlags().IntVar(&baseOptions.RedLimit, "reducer", 8, "sets the maximum parallel I/O, zero means no limit")
	cliCmd.PersistentFlags().StringArrayVar(&baseOptions.IndexImplems, "ximpl", nil, "list of non-default object storage index implementations")
	cliCmd.PersistentFlags().StringArrayVar(&baseOptions.ObsRegions, "obsrg", nil, "list of object storage regions")
	cliCmd.PersistentFlags().StringArrayVar(&baseOptions.ObsEndpoints, "obsep", nil, "list of object storage endpoints")
	cliCmd.PersistentFlags().StringArrayVar(&baseOptions.ObsContainers, "obsct", nil, "list of object storage containers")
	cliCmd.PersistentFlags().StringArrayVar(&baseOptions.ObsAccessKeys, "obsak", nil, "list of object storage access keys")
	cliCmd.PersistentFlags().StringArrayVar(&baseOptions.ObsSecretKeys, "obssk", nil, "list of object storage secret keys")
	cliCmd.PersistentFlags().StringVar(&baseOptions.TlsCert, "tlscrt", "", "certificate file on https server or untrusted CA on https client")
	cliCmd.PersistentFlags().BoolVar(&baseOptions.TlsNoCheck, "tlsnc", false, "no check of certificate by https client")
	cliCmd.PersistentFlags().StringVar(&baseOptions.HUser, "huser", "", "http client user")
	cliCmd.PersistentFlags().StringVar(&baseOptions.HPFile, "hpfile", "", "file containing the http client user password")
	cliCmd.PersistentFlags().BoolVar(&baseOptions.HPassword, "hpassword", false, "force http client user password prompt")
}
