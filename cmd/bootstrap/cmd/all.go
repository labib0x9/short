package cmd

import (
	"github.com/spf13/cobra"
)

var allCmd = &cobra.Command{
	Use:     "all",
	Aliases: []string{"all"},
	Short:   "initializes queue and sql migration",
	RunE:    allSetup,
}

func allSetup(cmd *cobra.Command, args []string) error {
	all = true
	err := setupDatabase()
	if err != nil {
		return err
	}
	return setupMessageQueue()
}

func init() {
	rootCmd.AddCommand(allCmd)
}
