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

func init() {
	dssMkCmd.Flags().StringVarP(&dssMkOptions.Size, "size", "s", "", "size is \"s\" for small, \"m\" for medium or \"l\" for large")
	dssCmd.AddCommand(dssMkCmd)
	dssMknsCmd.Flags().StringArrayVarP(&dssMknsOptions.Children, "children", "c", nil, "children")
	dssCmd.AddCommand(dssMknsCmd)
	dssUnlockCmd.Flags().BoolVar(&dssUnlockOptions.RepairIndex, "repair", false, "repair the index if persistent")
	dssUnlockCmd.Flags().BoolVar(&dssUnlockOptions.RepairReadOnly, "read", true, "don't repair, show diagnostic")
	dssCmd.AddCommand(dssUnlockCmd)
	dssCmd.AddCommand(dssCleanCmd)
}
