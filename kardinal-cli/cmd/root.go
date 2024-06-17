package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/compose-spec/compose-go/cli"
	"github.com/compose-spec/compose-go/types"
	"github.com/spf13/cobra"

	api "kardinal/cli-kontrol-api/api/golang/client"
	api_types "kardinal/cli-kontrol-api/api/golang/types"
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

		fmt.Println("Asking Kontrol to setup new dev flow")
		server(project)
		fmt.Println("done!")
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

	return project, nil
}

func server(project *types.Project) {
	ctx := context.Background()

	projectYAML, err := project.MarshalJSON()
	if err != nil {
		log.Fatal(err)
	}

	var dockerCompose map[string]interface{}
	err = json.Unmarshal(projectYAML, &dockerCompose)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	client, err := api.NewClientWithResponses("http://localhost:8080", api.WithHTTPClient(http.DefaultClient))
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	imageLocator := "someimage"
	serviceName := "kardinal"
	body := api_types.PostDevFlowJSONRequestBody{
		DockerCompose: &dockerCompose,
		ServiceName:   &serviceName,
		ImageLocator:  &imageLocator,
	}

	resp, err := client.PostDevFlowWithResponse(ctx, body)
	if err != nil {
		log.Fatalf("Failed to call greet: %v", err)
	}

	fmt.Printf("Response: %s\n", string(resp.Body))
}
