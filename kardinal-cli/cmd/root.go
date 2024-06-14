package cmd

import (
	"fmt"
	"log"

	"github.com/compose-spec/compose-go/cli"
	"github.com/compose-spec/compose-go/types"
	"github.com/spf13/cobra"
)

const projectName = "kardinal"

var composeFile string

var rootCmd = &cobra.Command{
	Use:   "kardinal",
	Short: "Setup new dev and test paths in your system",
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
	opts, err := cli.NewProjectOptions([]string{filename},
		cli.WithOsEnv,
		cli.WithDotEnv,
		cli.WithName(projectName),
	)
	if err != nil {
		return nil, err
	}

	project, err := cli.ProjectFromOptions(opts)
	if err != nil {
		return nil, err
	}

	projectYAML, err := project.MarshalYAML()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(projectYAML))

	return project, nil
}
