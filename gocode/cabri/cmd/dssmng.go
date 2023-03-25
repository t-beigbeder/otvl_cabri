package cmd

import (
	"fmt"

	"github.com/muesli/coral"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/cabriui"
)

var dssMkOptions cabriui.DSSMkOptions

var dssMkCmd = &coral.Command{
	Use:   "make <dss-type:/path/to/dss>",
	Short: "create a new DSS",
	Long:  `create a new DSS`,
	Args: func(cmd *coral.Command, args []string) error {
		if len(args) != 1 {
			cmd.UsageFunc()(cmd)
			return fmt.Errorf("a DSS specification must be provided")
		}
		dssType, _, err := cabriui.CheckDssSpec(args[0])
		if err != nil {
			cmd.UsageFunc()(cmd)
			return fmt.Errorf("%v\nsyntax: dss-type:/path/to/dss\nfor instance\n\tolf:/home/guest/olf_sample", err)
		}
		if dssType == "olf" && dssMkOptions.Size != "s" && dssMkOptions.Size != "m" && dssMkOptions.Size != "l" {
			cmd.UsageFunc()(cmd)
			return fmt.Errorf("incorrect size")
		}
		return nil
	},
	RunE: func(cmd *coral.Command, args []string) error {
		dssMkOptions.BaseOptions = baseOptions
		return cabriui.CLIRun[cabriui.DSSMkOptions, *cabriui.DSSMkVars](
			cmd.InOrStdin(), cmd.OutOrStdout(), cmd.ErrOrStderr(),
			dssMkOptions, args,
			cabriui.DSSMkStartup, cabriui.DSSMkShutdown)
	},
	SilenceUsage: true,
}

var dssMknsOptions = cabriui.DSSMknsOptions{Children: []string{}}

var dssMknsCmd = &coral.Command{
	Use:   "mkns",
	Short: "create a namespace",
	Long:  `create a namespace`,
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
		if _, err := cabriui.CheckUiACL(baseOptions.ACL); err != nil {
			return err
		}
		dssMknsOptions.BaseOptions = baseOptions
		return cabriui.CLIRun[cabriui.DSSMknsOptions, *cabriui.DSSMknsVars](
			cmd.InOrStdin(), cmd.OutOrStdout(), cmd.ErrOrStderr(),
			dssMknsOptions, args,
			cabriui.DSSMknsStartup, cabriui.DSSMknsShutdown)
	},
	SilenceUsage: true,
}

var dssUnlockOptions cabriui.DSSUnlockOptions

var dssUnlockCmd = &coral.Command{
	Use:   "unlock",
	Short: "unlock a DSS",
	Long:  `unlock a DSS`,
	Args: func(cmd *coral.Command, args []string) error {
		if len(args) != 1 {
			cmd.UsageFunc()(cmd)
			return fmt.Errorf("a DSS namespace must be provided")
		}
		_, _, err := cabriui.CheckDssSpec(args[0])
		if err != nil {
			cmd.UsageFunc()(cmd)
			return fmt.Errorf("%v\nsyntax: dss-type:/path/to/dss@path/in/dss\nfor instance\n\tfsy:/home/guest@Downloads", err)
		}
		return nil
	},
	RunE: func(cmd *coral.Command, args []string) error {
		dssUnlockOptions.BaseOptions = baseOptions
		return cabriui.CLIRun[cabriui.DSSUnlockOptions, *cabriui.DSSUnlockVars](
			cmd.InOrStdin(), cmd.OutOrStdout(), cmd.ErrOrStderr(),
			dssUnlockOptions, args,
			cabriui.DSSUnlockStartup, cabriui.DSSUnlockShutdown)
	},
	SilenceUsage: true,
}

var dssAuditOptions cabriui.DSSAuditOptions

var dssAuditCmd = &coral.Command{
	Use:   "audit",
	Short: "audit a DSS check files against index",
	Long:  `audit a DSS check files against index`,
	Args: func(cmd *coral.Command, args []string) error {
		if len(args) != 1 {
			cmd.UsageFunc()(cmd)
			return fmt.Errorf("a DSS namespace must be provided")
		}
		_, _, err := cabriui.CheckDssSpec(args[0])
		if err != nil {
			cmd.UsageFunc()(cmd)
			return fmt.Errorf("%v\nsyntax: dss-type:/path/to/dss@path/in/dss\nfor instance\n\tfsy:/home/guest@Downloads", err)
		}
		return nil
	},
	RunE: func(cmd *coral.Command, args []string) error {
		dssAuditOptions.BaseOptions = baseOptions
		return cabriui.CLIRun[cabriui.DSSAuditOptions, *cabriui.DSSAuditVars](
			cmd.InOrStdin(), cmd.OutOrStdout(), cmd.ErrOrStderr(),
			dssAuditOptions, args,
			cabriui.DSSAuditStartup, cabriui.DSSAuditShutdown)
	},
	SilenceUsage: true,
}

var dssScanOptions cabriui.DSSScanOptions

var dssScanCmd = &coral.Command{
	Use:   "scan",
	Short: "scan a DSS",
	Long:  `scan a DSS`,
	Args: func(cmd *coral.Command, args []string) error {
		if len(args) != 1 {
			cmd.UsageFunc()(cmd)
			return fmt.Errorf("a DSS namespace must be provided")
		}
		_, _, err := cabriui.CheckDssSpec(args[0])
		if err != nil {
			cmd.UsageFunc()(cmd)
			return fmt.Errorf("%v\nsyntax: dss-type:/path/to/dss@path/in/dss\nfor instance\n\tfsy:/home/guest@Downloads", err)
		}
		return nil
	},
	RunE: func(cmd *coral.Command, args []string) error {
		dssScanOptions.BaseOptions = baseOptions
		return cabriui.CLIRun[cabriui.DSSScanOptions, *cabriui.DSSScanVars](
			cmd.InOrStdin(), cmd.OutOrStdout(), cmd.ErrOrStderr(),
			dssScanOptions, args,
			cabriui.DSSScanStartup, cabriui.DSSScanShutdown)
	},
	SilenceUsage: true,
}

var dssReindexOptions cabriui.DSSReindexOptions

var dssReindexCmd = &coral.Command{
	Use:   "reindex",
	Short: "reindex a DSS",
	Long:  `reindex a DSS`,
	Args: func(cmd *coral.Command, args []string) error {
		if len(args) != 1 {
			cmd.UsageFunc()(cmd)
			return fmt.Errorf("a DSS namespace must be provided")
		}
		_, _, err := cabriui.CheckDssSpec(args[0])
		if err != nil {
			cmd.UsageFunc()(cmd)
			return fmt.Errorf("%v\nsyntax: dss-type:/path/to/dss@path/in/dss\nfor instance\n\tfsy:/home/guest@Downloads", err)
		}
		return nil
	},
	RunE: func(cmd *coral.Command, args []string) error {
		dssReindexOptions.BaseOptions = baseOptions
		return cabriui.CLIRun[cabriui.DSSReindexOptions, *cabriui.DSSReindexVars](
			cmd.InOrStdin(), cmd.OutOrStdout(), cmd.ErrOrStderr(),
			dssReindexOptions, args,
			cabriui.DSSReindexStartup, cabriui.DSSReindexShutdown)
	},
	SilenceUsage: true,
}

var dssLsHistoOptions cabriui.DSSLsHistoOptions

var dssLsHistoCmd = &coral.Command{
	Use:   "lshisto",
	Short: "list namespace or entry full history information",
	Long:  `list namespace or entry full history information`,
	Args: func(cmd *coral.Command, args []string) error {
		if len(args) != 1 {
			cmd.UsageFunc()(cmd)
			return fmt.Errorf("a DSS entry must be provided")
		}
		_, _, _, err := cabriui.CheckDssPath(args[0])
		if err != nil {
			cmd.UsageFunc()(cmd)
			return fmt.Errorf("%v\nsyntax: dss-type:/path/to/dss@path/in/dss\nfor instance\n\tolf:/home/guest/olf@Downloads", err)
		}
		return nil
	},
	RunE: func(cmd *coral.Command, args []string) error {
		dssLsHistoOptions.BaseOptions = baseOptions
		return cabriui.CLIRun[cabriui.DSSLsHistoOptions, *cabriui.DSSLsHistoVars](
			cmd.InOrStdin(), cmd.OutOrStdout(), cmd.ErrOrStderr(),
			dssLsHistoOptions, args,
			cabriui.DSSLsHistoStartup, cabriui.DSSLsHistoShutdown)
	},
	SilenceUsage: true,
}

var dssRmHistoOptions cabriui.DSSRmHistoOptions

var dssRmHistoCmd = &coral.Command{
	Use:   "rmhisto",
	Short: "removes history entries for a given time period",
	Long:  `removes history entries for a given time period`,
	Args: func(cmd *coral.Command, args []string) error {
		if len(args) != 1 {
			cmd.UsageFunc()(cmd)
			return fmt.Errorf("a DSS entry must be provided")
		}
		_, _, _, err := cabriui.CheckDssPath(args[0])
		if err != nil {
			cmd.UsageFunc()(cmd)
			return fmt.Errorf("%v\nsyntax: dss-type:/path/to/dss@path/in/dss\nfor instance\n\tolf:/home/guest/olf@Downloads", err)
		}
		return nil
	},
	RunE: func(cmd *coral.Command, args []string) error {
		dssRmHistoOptions.BaseOptions = baseOptions
		if _, err := cabriui.CheckTimeStamp(dssRmHistoOptions.StartTime); err != nil {
			return err
		}
		if _, err := cabriui.CheckTimeStamp(dssRmHistoOptions.EndTime); err != nil {
			return err
		}
		return cabriui.CLIRun[cabriui.DSSRmHistoOptions, *cabriui.DSSRmHistoVars](
			cmd.InOrStdin(), cmd.OutOrStdout(), cmd.ErrOrStderr(),
			dssRmHistoOptions, args,
			cabriui.DSSRmHistoStartup, cabriui.DSSRmHistoShutdown)
	},
	SilenceUsage: true,
}

var dssCleanOptions cabriui.DSSCleanOptions

var dssCleanCmd = &coral.Command{
	Use:   "clean",
	Short: "completely clean an OBS DSS",
	Long:  `completely clean an OBS DSS`,
	Args: func(cmd *coral.Command, args []string) error {
		if len(args) != 1 {
			cmd.UsageFunc()(cmd)
			return fmt.Errorf("a DSS namespace must be provided")
		}
		_, _, err := cabriui.CheckDssSpec(args[0])
		if err != nil {
			cmd.UsageFunc()(cmd)
			return fmt.Errorf("%v\nsyntax: dss-type:/path/to/dss@path/in/dss\nfor instance\n\tfsy:/home/guest@Downloads", err)
		}
		return nil
	},
	RunE: func(cmd *coral.Command, args []string) error {
		dssCleanOptions.BaseOptions = baseOptions
		return cabriui.CLIRun[cabriui.DSSCleanOptions, *cabriui.DSSCleanVars](
			cmd.InOrStdin(), cmd.OutOrStdout(), cmd.ErrOrStderr(),
			dssCleanOptions, args,
			cabriui.DSSCleanStartup, cabriui.DSSCleanShutdown)
	},
	Hidden:       true,
	SilenceUsage: true,
}

var dssConfigOptions cabriui.DSSConfigOptions

var dssConfigCmd = &coral.Command{
	Use:   "config",
	Short: "updates and/or displays the DSS configuration",
	Long:  `updates and/or displays the DSS configuration`,
	Args: func(cmd *coral.Command, args []string) error {
		if len(args) != 1 {
			cmd.UsageFunc()(cmd)
			return fmt.Errorf("a DSS specification must be provided")
		}
		_, _, err := cabriui.CheckDssSpec(args[0])
		if err != nil {
			cmd.UsageFunc()(cmd)
			return fmt.Errorf("%v\nsyntax: dss-type:/path/to/dss\nfor instance\n\tolf:/home/guest/olf", err)
		}
		return nil
	},
	RunE: func(cmd *coral.Command, args []string) error {
		dssConfigOptions.BaseOptions = baseOptions
		return cabriui.CLIRun[cabriui.DSSConfigOptions, *cabriui.DSSConfigVars](
			cmd.InOrStdin(), cmd.OutOrStdout(), cmd.ErrOrStderr(),
			dssConfigOptions, args,
			cabriui.DSSConfigStartup, cabriui.DSSConfigShutdown)
	},
	SilenceUsage: true,
}

func init() {
	cliCmd.AddCommand(dssCmd)
	dssMkCmd.Flags().StringVarP(&dssMkOptions.Size, "size", "s", "", "size is \"s\" for small, \"m\" for medium or \"l\" for large")
	dssCmd.AddCommand(dssMkCmd)
	dssMknsCmd.Flags().StringArrayVarP(&dssMknsOptions.Children, "children", "c", nil, "children")
	dssCmd.AddCommand(dssMknsCmd)
	dssUnlockCmd.Flags().BoolVar(&dssUnlockOptions.RepairIndex, "repair", false, "repair the index if persistent")
	dssUnlockCmd.Flags().BoolVar(&dssUnlockOptions.RepairReadOnly, "read", true, "don't repair, show diagnostic")
	dssCmd.AddCommand(dssUnlockCmd)
	dssCmd.AddCommand(dssAuditCmd)
	dssCmd.AddCommand(dssScanCmd)
	dssCmd.AddCommand(dssReindexCmd)
	dssLsHistoCmd.Flags().BoolVarP(&dssLsHistoOptions.Recursive, "recursive", "r", false, "recursively list subnamespaces information")
	dssLsHistoCmd.Flags().BoolVarP(&dssLsHistoOptions.Sorted, "sorted", "s", false, "sort entries by name")
	dssCmd.AddCommand(dssLsHistoCmd)
	dssRmHistoCmd.Flags().BoolVarP(&dssRmHistoOptions.Recursive, "recursive", "r", false, "recursively remove the history of all namespace children")
	dssRmHistoCmd.Flags().BoolVarP(&dssRmHistoOptions.DryRun, "dryrun", "d", false, "don't remove the history, just report work to be done")
	dssRmHistoCmd.Flags().StringVar(&dssRmHistoOptions.StartTime, "st", "", "inclusive index time above which entries must be removed, default to all past entries")
	dssRmHistoCmd.Flags().StringVar(&dssRmHistoOptions.EndTime, "et", "", "the inclusive index time below which entries must be removed, default to all future entries")
	dssCmd.AddCommand(dssRmHistoCmd)
	dssCmd.AddCommand(dssConfigCmd)
	dssCmd.AddCommand(dssCleanCmd)
	dssCmd.AddCommand(dssConfigCmd)
}
