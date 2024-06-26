package cmd

import (
	"fmt"

	"github.com/muesli/coral"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/cabriui"
)

var syncOptions cabriui.SyncOptions

var syncCmd = &coral.Command{
	Use:   "sync",
	Short: "synchronizes two DSS subtrees",
	Long:  `synchronizes two DSS subtrees`,
	Args: func(cmd *coral.Command, args []string) error {
		returnUsageAndErr := func(err error) error {
			cmd.UsageFunc()(cmd)
			return err
		}
		checkArg := func(arg string) error {
			if _, _, _, err := cabriui.CheckDssPath(arg); err != nil {
				return fmt.Errorf(`
%v
DSS syntax is: dss-type:/path/to/dss@path/in/dss
for instance
	fsy:/home/guest@Downloads`,
					err,
				)
			}
			return nil
		}
		if len(args) != 2 {
			return returnUsageAndErr(fmt.Errorf("two DSS namespaces must be provided"))
		}
		if err := checkArg(args[0]); err != nil {
			return returnUsageAndErr(err)
		}
		if err := checkArg(args[1]); err != nil {
			return returnUsageAndErr(err)
		}
		return nil
	},
	RunE: func(cmd *coral.Command, args []string) error {
		if _, err := cabriui.CheckUiACL(baseOptions.ACL); err != nil {
			return err
		}
		if _, err := cabriui.CheckUiACL(syncOptions.LeftACL); err != nil {
			return err
		}
		baseOptions.LeftUsers = syncOptions.LeftUsers
		baseOptions.LeftACL = syncOptions.LeftACL
		syncOptions.BaseOptions = baseOptions
		if _, err := cabriui.CheckTimeStamp(syncOptions.LeftTime); err != nil {
			return err
		}
		if _, err := cabriui.CheckTimeStamp(syncOptions.RightTime); err != nil {
			return err
		}
		return cabriui.CLIRun[cabriui.SyncOptions, *cabriui.SyncVars](
			cmd.InOrStdin(), cmd.OutOrStdout(), cmd.ErrOrStderr(),
			syncOptions, args,
			cabriui.SyncStartup, cabriui.SyncShutdown)
	},
	SilenceUsage: true,
}

func init() {
	cliCmd.AddCommand(syncCmd)
	syncCmd.Flags().BoolVarP(&syncOptions.Recursive, "recursive", "r", false, "synchronize sub-namespaces content recursively")
	syncCmd.Flags().BoolVarP(&syncOptions.DryRun, "dryrun", "d", false, "don't synchronize, just report work to be done")
	syncCmd.Flags().BoolVarP(&syncOptions.BiDir, "bidir", "b", false, "bidirectional synchronization, the latest modified content wins")
	syncCmd.Flags().BoolVarP(&syncOptions.KeepContent, "keep", "k", false, "don't remove content deleted from one side in other side")
	syncCmd.Flags().BoolVarP(&syncOptions.NoCh, "nocheck", "n", false, "don't evaluate checksum when not available, compare content's size and modification time")
	syncCmd.Flags().StringArrayVar(&syncOptions.Exclude, "excl", nil, "list of regular expression patterns to exclude from sync")
	syncCmd.Flags().StringArrayVar(&syncOptions.ExcludeFrom, "exclfile", nil, "list of files containing regular expression patterns to exclude from sync")
	syncCmd.Flags().BoolVar(&syncOptions.Summary, "summary", false, "only displays synchronization summary")
	syncCmd.Flags().BoolVar(&syncOptions.DisplayRight, "dispright", false, "display right entries in report even if equal to left")
	syncCmd.Flags().BoolVarP(&syncOptions.Verbose, "verbose", "v", false, "display synchronization statistics")
	syncCmd.Flags().IntVar(&syncOptions.VerboseLevel, "debug", 0, "display synchronization debug messages if level >= 2")
	syncCmd.Flags().StringVar(&syncOptions.LeftTime, "lefttime", "", "upper time of entries retrieved in left historized DSS")
	syncCmd.Flags().StringVar(&syncOptions.RightTime, "righttime", "", "upper time of entries retrieved in right historized DSS")
	syncCmd.Flags().BoolVar(&syncOptions.NoACL, "noacl", false, "don't check ACL")
	syncCmd.Flags().StringArrayVar(&syncOptions.MapACL, "macl", nil, "list of ACL user mapping <left-user:right-user> items")
	syncCmd.PersistentFlags().StringArrayVar(&syncOptions.LeftUsers, "leftuser", nil, "list of ACL users for left-side retrieval")
	syncCmd.PersistentFlags().StringArrayVar(&syncOptions.LeftACL, "leftacl", nil, "list of ACL <user:rights> items (defaults to rw) for left-side creation and update")
}
