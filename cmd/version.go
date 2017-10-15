package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/steinarvk/heisenlisp/version"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Shows version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Version:       ", version.VersionString)
		fmt.Println("Commit:        ", version.CommitHash)
		fmt.Println("Build time:    ", version.BuildTimestampISO8601)
		fmt.Println("Build machine: ", version.BuildMachineInfo)
		fmt.Println("Go version:    ", version.GoVersion)
	},
}

func init() {
	RootCmd.AddCommand(versionCmd)
}
