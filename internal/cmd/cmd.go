package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// Build info set by ldflags
var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

func Run(stderr *os.File, stdout *os.File, args []string, stdin *os.File) int {
	rootCmd := cobra.Command{
		Use:     "gofast",
		Short:   "gofast is a CLI tool to generate Go REST APIs quickly",
		Version: fmt.Sprintf("%s (commit: %s, built: %s)", Version, Commit, Date),
	}

	rootCmd.AddCommand(generateCmd)

	rootCmd.SetOut(stdout)
	rootCmd.SetErr(stderr)
	rootCmd.SetArgs(args)

	if err := rootCmd.Execute(); err != nil {
		return 1
	}

	return 0
}

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate a new Go REST API project",
	RunE: func(cmd *cobra.Command, args []string) error {
		wd, err := os.Getwd()
		if err != nil {
			return err
		}

		if err := generate(wd); err != nil {
			return err
		}
		return nil
	},
}
