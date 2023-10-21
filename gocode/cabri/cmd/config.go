package cmd

import (
	"fmt"

	"github.com/muesli/coral"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/cabriui"
)

var configOptions cabriui.ConfigOptions

var configCmd = &coral.Command{
	Use:   "config [args...]",
	Short: "manage application configuration",
	Long:  `manage application configuration`,
	Args: func(cmd *coral.Command, args []string) error {
		return nil
	},
	RunE: func(cmd *coral.Command, args []string) error {
		configOptions.BaseOptions = baseOptions
		var (
			ff  string
			err error
		)
		if ff, err = cabriui.MutualExcludeFlags(
			[]string{"encrypt", "decrypt", "dump", "gen", "get", "put", "remove"},
			configOptions.Encrypt, configOptions.Decrypt, configOptions.Dump,
			configOptions.Gen, configOptions.Get, configOptions.Put, configOptions.Remove); err != nil {
			return err
		}
		if ff == "" {
			return fmt.Errorf("at least one operation option must be given with the config command")
		}
		return cabriui.CLIRun[cabriui.ConfigOptions, *cabriui.ConfigVars](
			cmd.InOrStdin(), cmd.OutOrStdout(), cmd.ErrOrStderr(),
			configOptions, args,
			cabriui.ConfigStartup, cabriui.ConfigShutdown)
	},
	SilenceUsage: true,
}

func init() {
	cliCmd.AddCommand(configCmd)
	configCmd.Flags().BoolVarP(&configOptions.Encrypt, "encrypt", "e", false, "encrypts the configuration file with master password")
	configCmd.Flags().BoolVarP(&configOptions.Decrypt, "decrypt", "d", false, "decrypts the configuration file with master password")
	configCmd.Flags().BoolVarP(&configOptions.Dump, "dump", "", false, "dumps the configuration file")
	configCmd.Flags().BoolVarP(&configOptions.Gen, "gen", "", false, "generate a new identity for one or several aliases")
	configCmd.Flags().BoolVarP(&configOptions.Get, "get", "", false, "display an identity for one or several aliases")
	configCmd.Flags().BoolVarP(&configOptions.Put, "put", "", false, "<alias> <pkey> [<secret>] import or update an identity for an alias, secret may be unknown")
	configCmd.Flags().BoolVarP(&configOptions.Remove, "remove", "", false, "remove an identity alias")
}
