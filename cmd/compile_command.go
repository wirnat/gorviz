package cmd

import (
	"fmt"
	"os"

	"github.com/wirnat/gorviz/internal"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var compileCmd = &cobra.Command{
	Use:   "compile",
	Short: "Compiles the YAML schema into a static HTML documentation with interactive ERD",
	Long: `The 'compile' command reads the generated YAML file and transforms it into a 
comprehensive static HTML documentation file. This includes both a detailed 
browsable list of models and an interactive Entity-Relationship Diagram (ERD).`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Compiling YAML schema to HTML documentation and ERD...")

		// Read schema.yaml
		yamlFile, err := os.ReadFile("schema.yaml")
		if err != nil {
			fmt.Printf("Error reading schema.yaml. Make sure you've run 'init' command first: %v\n", err)
			os.Exit(1)
		}

		var schema internal.Schema
		err = yaml.Unmarshal(yamlFile, &schema)
		if err != nil {
			fmt.Printf("Error unmarshalling schema.yaml: %v\n", err)
			os.Exit(1)
		}

		// Generate HTML content
		htmlContent, err := generateStaticHTML(&schema)
		if err != nil {
			fmt.Printf("Error generating HTML: %v\n", err)
			os.Exit(1)
		}

		outputFile := "gorviz.html"
		err = os.WriteFile(outputFile, []byte(htmlContent), 0644)
		if err != nil {
			fmt.Printf("Error writing %s: %v\n", outputFile, err)
			os.Exit(1)
		}

		fmt.Printf("Successfully generated %s. You can now open it in your browser.\n", outputFile)
	},
}

func init() {
	rootCmd.AddCommand(compileCmd)
}
