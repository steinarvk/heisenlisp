package cmd

import (
	"log"

	"github.com/spf13/cobra"
)

var (
	verbose *bool
)

func init() {
	verbose = RootCmd.PersistentFlags().Bool("verbose", false, "increase verbosity")
}

var RootCmd = &cobra.Command{
	Use:   "heisenlisp",
	Short: "Heisenlisp is a Lisp with features for uncertainty",
	Long:  "Heisenlisp is a Lisp with features for uncertainty",
	Run: func(cmd *cobra.Command, args []string) {
		if err := runREPL(); err != nil {
			log.Fatal(err)
		}
	},
}
