package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/steinarvk/heisenlisp/builtin"
	"github.com/steinarvk/heisenlisp/code"
)

var evalCmd = &cobra.Command{
	Use:   "eval [expr...]",
	Short: "Evaluates a single expression",
	Long: `eval evaluates a single expression given on the command line. It writes
the result to stdout. For example:

  $ heisenlisp eval (+ 2 2)
	4`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		root := builtin.NewRootEnv()
		for _, arg := range args {
			value, err := code.Run(root, "<cmdline expr>", []byte(arg))
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(value)
		}
	},
}

func init() {
	RootCmd.AddCommand(evalCmd)
}
