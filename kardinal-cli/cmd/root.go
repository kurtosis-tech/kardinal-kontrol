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
	Short: "Kardinal CLI to manage deployment flows",
}

var flowCmd = &cobra.Command{
	Use:   "flow",
	Short: "Manage deployment flows",
}

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy services",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		services, err := parseComposeFile(composeFile)
		if err != nil {
			log.Fatalf("Error loading compose file: %v", err)
		}
		deploy(services)
	},
}

var createCmd = &cobra.Command{
	Use:   "create [service name] [image name]",
	Short: "Create a new service in development mode",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		serviceName, imageName := args[0], args[1]
		services, err := parseComposeFile(composeFile)
		if err != nil {
			log.Fatalf("Error loading compose file: %v", err)
		}

		fmt.Printf("Creating service %s with image %s in development mode...\n", serviceName, imageName)
		createDevFlow(services, imageName, serviceName)
	},
}

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete services",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		services, err := parseComposeFile(composeFile)
		if err != nil {
			log.Fatalf("Error loading compose file: %v", err)
		}
		deleteFlow(services)

		fmt.Print("Deleting dev flow")
	},
}

func init() {
	rootCmd.AddCommand(flowCmd)
	rootCmd.AddCommand(deployCmd)
	flowCmd.AddCommand(createCmd, deleteCmd)

	flowCmd.PersistentFlags().StringVarP(&composeFile, "docker-compose", "d", "", "Path to the Docker Compose file")
	flowCmd.MarkPersistentFlagRequired("docker-compose")
	deployCmd.PersistentFlags().StringVarP(&composeFile, "docker-compose", "d", "", "Path to the Docker Compose file")
	deployCmd.MarkPersistentFlagRequired("docker-compose")
}

func Execute() error {
	return rootCmd.Execute()
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

func parseComposeFile(composeFile string) ([]types.ServiceConfig, error) {
	project, err := loadComposeFile(composeFile)
	if err != nil {
		log.Fatalf("Error loading compose file: %v", err)
		return nil, err
	}

	fmt.Println("Services in the Docker Compose file:")
	for _, service := range project.Services {
		fmt.Println(service.Name)
	}

	projectYAML, err := project.MarshalJSON()
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	var dockerCompose map[string]interface{}
	err = json.Unmarshal(projectYAML, &dockerCompose)
	if err != nil {
		log.Fatalf("error: %v", err)
		return nil, err
	}

	return project.Services, nil
}

func createDevFlow(services []types.ServiceConfig, imageLocator, serviceName string) {
	ctx := context.Background()

	body := api_types.PostFlowCreateJSONRequestBody{
		DockerCompose: &services,
		ServiceName:   &serviceName,
		ImageLocator:  &imageLocator,
	}
	client, err := api.NewClientWithResponses("http://localhost:8080", api.WithHTTPClient(http.DefaultClient))
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	resp, err := client.PostFlowCreateWithResponse(ctx, body)
	if err != nil {
		log.Fatalf("Failed to create dev flow: %v", err)
	}

	fmt.Printf("Response: %s\n", string(resp.Body))
}

func deploy(services []types.ServiceConfig) {
	ctx := context.Background()

	body := api_types.PostDeployJSONRequestBody{
		DockerCompose: &services,
	}
	client, err := api.NewClientWithResponses("http://localhost:8080", api.WithHTTPClient(http.DefaultClient))
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	resp, err := client.PostDeployWithResponse(ctx, body)
	if err != nil {
		log.Fatalf("Failed to deploy: %v", err)
	}

	fmt.Printf("Response: %s\n", string(resp.Body))
}

func deleteFlow(services []types.ServiceConfig) {
	ctx := context.Background()

	body := api_types.PostFlowDeleteJSONRequestBody{
		DockerCompose: &services,
	}
	client, err := api.NewClientWithResponses("http://localhost:8080", api.WithHTTPClient(http.DefaultClient))
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	resp, err := client.PostFlowDeleteWithResponse(ctx, body)
	if err != nil {
		log.Fatalf("Failed to delete flow: %v", err)
	}

	fmt.Printf("Response: %s\n", string(resp.Body))
}
