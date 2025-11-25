package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/wirnat/gorviz/parser"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var initCmd = &cobra.Command{
	Use:   "init [folder_location]",
	Short: "Scans GORM models in the specified folder and generates a YAML schema",
	Long: `The 'init' command scans all .go files in the given folder location,
parses GORM model structs, extracts relationships based on GORM tags,
and generates a YAML file representing the models and their connections.`,
	Args: cobra.ExactArgs(1), // Requires exactly one argument: the folder location
	Run: func(cmd *cobra.Command, args []string) {
		folderLocation := args[0]
		fmt.Printf("Scanning GORM models in: %s\n", folderLocation)

		absPath, err := filepath.Abs(folderLocation)
		if err != nil {
			fmt.Printf("Error getting absolute path: %v\n", err)
			os.Exit(1)
		}

		schema, err := parser.ParseGormModels(absPath)
		if err != nil {
			fmt.Printf("Error parsing GORM models: %v\n", err)
			os.Exit(1)
		}

		outputYAML, err := yaml.Marshal(schema)
		if err != nil {
			fmt.Printf("Error marshalling schema to YAML: %v\n", err)
			os.Exit(1)
		}

		err = os.WriteFile("schema.yaml", outputYAML, 0644)
		if err != nil {
			fmt.Printf("Error writing schema.yaml: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Successfully generated schema.yaml")
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
	// Here you can define flags specific to the 'init' command
}
