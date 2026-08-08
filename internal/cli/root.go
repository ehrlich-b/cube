package cli

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "cube",
	Short: "A flexible Rubik's cube solver",
	Long: `Cube is a flexible Rubik's cube solver that supports multiple dimensions
and solving algorithms.`,
	Version: "1.0.0",
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(solveCmd)
	rootCmd.AddCommand(twistCmd)
	rootCmd.AddCommand(verifyCmd)
	rootCmd.AddCommand(showCmd)
	rootCmd.AddCommand(lookupCmd)
	// optimizeCmd and findCmd register themselves in their own init() functions
	// (optimize.go, find.go); registering them here too would duplicate them.
}
