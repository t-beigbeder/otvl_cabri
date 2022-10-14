package cmd

import (
	"os"

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
