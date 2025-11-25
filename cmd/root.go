package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "gorviz",
	Short: "A tool to visualize GORM models and their relationships",
	Long: `gorviz is a CLI tool that helps you to:
1. Scan your Go project for GORM models and extract their structure and relationships (based on gorm tags).
2. Compile the extracted data into an interactive HTML visualization.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}

func init() {
	// Here you will initialize your configuration and flags
	// and connect them to the root command.
	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.gorm_visualization.yaml)")
	// rootCmd.PersistentFlags().BoolVarP(&toggle, "toggle", "t", false, "Help message for toggle")
}
