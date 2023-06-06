package cmd

import (
	"os"

	"github.com/t-beigbeder/otvl_cabri/gocode/packages/cabridss"

	"github.com/muesli/coral"
)

var rootCmd = &coral.Command{
	Use:   "cabri",
	Short: "Cabri CLI or batch services",
	Long:  `Cabri CLI or batch services`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(-1)
	}
}

var versionCmd = &coral.Command{
	Use:   "version",
	Short: "displays version",
	Long:  "displays version",
	RunE: func(cmd *coral.Command, args []string) error {
		println(cabridss.CabriVersion)
		return nil
	},
	SilenceUsage: true,
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
