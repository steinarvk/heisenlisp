package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/steinarvk/heisenlisp/builtin"
	"github.com/steinarvk/heisenlisp/code"
)

var runCmd = &cobra.Command{
	Use:   "run [filename.hlisp]",
	Short: "Runs a file",
	Long:  `run runs a Heisenlisp script file and prints the last expression.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		root := builtin.NewRootEnv()
		value, err := code.RunFile(root, args[0])
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(value)
	},
}

func init() {
	RootCmd.AddCommand(runCmd)
}
