package cmd

import (
	"fmt"
	"log"

	"github.com/compose-spec/compose-go/cli"
	"github.com/compose-spec/compose-go/types"
	"github.com/spf13/cobra"
)

var composeFile string

var rootCmd = &cobra.Command{
	Use:   "docker-compose-parser",
	Short: "Parses a Docker Compose file and lists services",
	Run: func(cmd *cobra.Command, args []string) {
		project, err := loadComposeFile(composeFile)
		if err != nil {
			log.Fatalf("Error loading compose file: %v", err)
		}

		fmt.Println("Services in the Docker Compose file:")
		for _, service := range project.Services {
			fmt.Println(service.Name)
		}
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.Flags().StringVarP(&composeFile, "docker-compose", "d", "", "Path to the Docker Compose file")
	rootCmd.MarkFlagRequired("docker-compose")
}

func loadComposeFile(filename string) (*types.Project, error) {
	opts, err := cli.NewProjectOptions([]string{filename})
	if err != nil {
		return nil, err
	}

	project, err := cli.ProjectFromOptions(opts)
	if err != nil {
		return nil, err
	}

	return project, nil
}
