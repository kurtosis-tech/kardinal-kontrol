package plugins

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"kardinal.kontrol-service/database"

	appv1 "k8s.io/api/apps/v1"
)

const (
	// <flow id>-<service id>-<plugin idx>
	pluginIdFmtStr = "%s-%s-%d"
	// <flow id>-<service1 id>,<service2 id>,<service3 id>
	pluginIdFmtStr2 = "%s-%s"
)

type PluginRunner struct {
	gitPluginProvider GitPluginProvider

	tenantId string

	db *database.Db
}

func NewPluginRunner(gitPluginProvider GitPluginProvider, tenantId string, db *database.Db) *PluginRunner {
	return &PluginRunner{
		gitPluginProvider: gitPluginProvider,
		tenantId:          tenantId,
		db:                db,
	}
}

func (pr *PluginRunner) CreateFlow(pluginUrl string, serviceSpecs []corev1.ServiceSpec, deploymentSpecs []appv1.DeploymentSpec, flowUuid string, arguments map[string]string) ([]appv1.DeploymentSpec, string, error) {
	repoPath, err := pr.getOrCloneRepo(pluginUrl)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get or clone repository: %v", err)
	}

	serviceSpecsJSON, err := json.Marshal(serviceSpecs)
	if err != nil {
		return nil, "", fmt.Errorf("failed to marshal service specs: %v", err)
	}
	serviceSpecsJSONStr := base64.StdEncoding.EncodeToString(serviceSpecsJSON)

	deploymentSpecsJSON, err := json.Marshal(deploymentSpecs)
	if err != nil {
		return nil, "", fmt.Errorf("failed to marshal deployment specs: %v", err)
	}
	deploymentSpecsJSONStr := base64.StdEncoding.EncodeToString(deploymentSpecsJSON)

	result, err := runPythonCreateFlow(repoPath, serviceSpecsJSONStr, deploymentSpecsJSONStr, flowUuid, arguments)
	if err != nil {
		return nil, "", err
	}

	var resultMap map[string]json.RawMessage
	err = json.Unmarshal([]byte(result), &resultMap)
	if err != nil {
		return nil, "", fmt.Errorf("failed to parse result: %v", err)
	}

	var newDeploymentSpecs []appv1.DeploymentSpec
	err = json.Unmarshal(resultMap["deployment_specs"], &newDeploymentSpecs)
	if err != nil {
		return nil, "", fmt.Errorf("failed to unmarshal deployment spec: %v", err)
	}

	configMapJSON := resultMap["config_map"]
	var configMap map[string]interface{}
	err = json.Unmarshal(configMapJSON, &configMap)
	if err != nil {
		return nil, "", fmt.Errorf("invalid config map: %v", err)
	}

	configMapBytes, err := json.Marshal(configMap)
	if err != nil {
		return nil, "", fmt.Errorf("failed to re-marshal config map: %v", err)
	}

	logrus.Infof("Storing config map for plugin called with uuid '%v':\n %s\n...", flowUuid, string(configMapBytes))
	_, err = pr.db.CreatePluginConfig(flowUuid, string(configMapBytes), pr.tenantId)
	if err != nil {
		return nil, "", fmt.Errorf("failed to store the config map: %v", err)
	}

	return newDeploymentSpecs, string(configMapBytes), nil
}

func (pr *PluginRunner) DeleteFlow(pluginUrl, flowUuid string) error {
	repoPath, err := pr.getOrCloneRepo(pluginUrl)
	if err != nil {
		return fmt.Errorf("failed to get or clone repository: %v", err)
	}

	configMap, err := pr.getConfigForFlow(flowUuid)
	if err != nil {
		return err
	}

	_, err = runPythonDeleteFlow(repoPath, configMap, flowUuid)
	if err != nil {
		return err
	}

	err = pr.db.DeletePluginConfig(pr.tenantId, flowUuid)
	if err != nil {
		return err
	}

	return nil
}

// TODO remove this after the DeleteDevFlow refactor
func GetPluginId(flowId, serviceId string, pluginIdx int) string {
	return fmt.Sprintf(pluginIdFmtStr, flowId, serviceId, pluginIdx)
}

// TODO rename it to the original name
func GetPluginId2(flowId string, serviceIds []string) string {
	serviceIdsStr := strings.Join(serviceIds, ",")
	return fmt.Sprintf(pluginIdFmtStr2, flowId, serviceIdsStr)
}

func (pr *PluginRunner) getConfigForFlow(flowUuid string) (string, error) {
	pluginConfig, err := pr.db.GetPluginConfigByFlowID(pr.tenantId, flowUuid)
	if err != nil {
		return "", err
	}
	if pluginConfig == nil {
		return "", fmt.Errorf("no config map found for flow UUID: %s", flowUuid)
	}
	return pluginConfig.Config, nil
}

func runPythonCreateFlow(repoPath, serviceSpecsJSONStr, deploymentSpecsJSONStr, flowUuid string, arguments map[string]string) (string, error) {
	scriptPath := filepath.Join(repoPath, "main.py")

	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		return "", fmt.Errorf("main.py not found in the repository")
	}

	venvPath := filepath.Join(repoPath, "venv")
	if err := createVirtualEnv(venvPath); err != nil {
		return "", fmt.Errorf("failed to create virtual environment: %v", err)
	}

	requirementsPath := filepath.Join(repoPath, "requirements.txt")
	if _, err := os.Stat(requirementsPath); err == nil {
		if err := installDependencies(venvPath, requirementsPath); err != nil {
			return "", fmt.Errorf("failed to install dependencies: %v", err)
		}
	}

	// Convert arguments to JSON, then encode it for Python
	argsJSON, err := json.Marshal(arguments)
	if err != nil {
		return "", fmt.Errorf("failed to marshal arguments: %v", err)
	}
	argJsonStr := base64.StdEncoding.EncodeToString(argsJSON)

	tempResultFile, err := os.CreateTemp("", "result_*.json")
	if err != nil {
		return "", fmt.Errorf("failed to create temporary result file: %v", err)
	}
	defer os.Remove(tempResultFile.Name())

	tempScript := fmt.Sprintf(`
import sys
import json
import inspect
import base64
sys.path.append("%s")
import main

service_specs_json = base64.b64decode('%s').decode('utf-8')
service_specs = json.loads(service_specs_json)
deployment_specs_json = base64.b64decode('%s').decode('utf-8')
deployment_specs = json.loads(deployment_specs_json)
flow_uuid = %q
args_json = base64.b64decode('%s').decode('utf-8')
args = json.loads(args_json)

# Get the signature of the create_flow function
sig = inspect.signature(main.create_flow)

# Prepare kwargs based on the function signature
kwargs = {}
for param in sig.parameters.values():
    if param.name == 'service_specs':
        kwargs['service_specs'] = service_specs
    elif param.name == 'deployment_specs':
        kwargs['deployment_specs'] = deployment_specs
    elif param.name == 'flow_uuid':
        kwargs['flow_uuid'] = flow_uuid
    elif param.name in args:
        kwargs[param.name] = args[param.name]
    elif param.default is not param.empty:
        kwargs[param.name] = param.default
    else:
        print(f"Warning: Required parameter {param.name} not provided")

# Call create_flow with the prepared kwargs
result = main.create_flow(**kwargs)

# Write the result to a temporary file
with open('%s', 'w') as f:
    json.dump(result, f)
`, repoPath, serviceSpecsJSONStr, deploymentSpecsJSONStr, flowUuid, argJsonStr, tempResultFile.Name())

	if err := executePythonScript(venvPath, repoPath, tempScript); err != nil {
		return "", err
	}

	// Read the result from the temporary file
	resultBytes, err := os.ReadFile(tempResultFile.Name())
	if err != nil {
		return "", fmt.Errorf("failed to read result file: %v", err)
	}

	return string(resultBytes), nil
}

func runPythonDeleteFlow(repoPath, configMap, flowUuid string) (string, error) {

	scriptPath := filepath.Join(repoPath, "main.py")

	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		return "", fmt.Errorf("main.py not found in the repository")
	}

	venvPath := filepath.Join(repoPath, "venv")
	if err := createVirtualEnv(venvPath); err != nil {
		return "", fmt.Errorf("failed to create virtual environment: %v", err)
	}

	requirementsPath := filepath.Join(repoPath, "requirements.txt")
	if _, err := os.Stat(requirementsPath); err == nil {
		if err := installDependencies(venvPath, requirementsPath); err != nil {
			return "", fmt.Errorf("failed to install dependencies: %v", err)
		}
	}

	tempResultFile, err := os.CreateTemp("", "result_*.json")
	if err != nil {
		return "", fmt.Errorf("failed to create temporary result file: %v", err)
	}
	defer os.Remove(tempResultFile.Name())

	tempScript := fmt.Sprintf(`
import sys
import json
import inspect
sys.path.append("%s")
import main

config_map = %s
flow_uuid = %q
sig = inspect.signature(main.delete_flow)

kwargs = {}
for param in sig.parameters.values():
    if param.name == 'flow_uuid':
        kwargs['flow_uuid'] = flow_uuid
    elif param.name == 'config_map':
        kwargs['config_map'] = config_map
    elif param.default is not param.empty:
        kwargs[param.name] = param.default
    else:
        print(f"Warning: Required parameter {param.name} not provided")

# Call delete_flow with the prepared kwargs
result = main.delete_flow(**kwargs)

# Write the result to a temporary file
with open('%s', 'w') as f:
    json.dump(result, f)
`, repoPath, configMap, flowUuid, tempResultFile.Name())

	if err := executePythonScript(venvPath, repoPath, tempScript); err != nil {
		return "", err
	}

	// Read the result from the temporary file
	resultBytes, err := os.ReadFile(tempResultFile.Name())
	if err != nil {
		return "", fmt.Errorf("failed to read result file: %v", err)
	}

	return string(resultBytes), nil
}

func executePythonScript(venvPath, repoPath, scriptContent string) error {
	tempFile, err := os.CreateTemp("", "temp_script_*.py")
	if err != nil {
		return fmt.Errorf("failed to create temporary script: %v", err)
	}
	defer os.Remove(tempFile.Name())

	if _, err := tempFile.WriteString(scriptContent); err != nil {
		return fmt.Errorf("failed to write temporary script: %v", err)
	}
	tempFile.Close()

	cmd := exec.Command(filepath.Join(venvPath, "bin", "python"), tempFile.Name())
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to run Python script: %v\nOutput: %s", err, output)
	}

	return nil
}

func (pr *PluginRunner) getOrCloneRepo(repoURL string) (string, error) {
	if !strings.HasPrefix(repoURL, "https://") {
		repoURL = "https://" + repoURL
	}
	if !strings.HasSuffix(repoURL, ".git") {
		repoURL = repoURL + ".git"
	}

	parts := strings.Split(repoURL, "/")
	repoName := parts[len(parts)-1]

	tempDir := filepath.Join(os.TempDir(), "go-python-plugins")
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create temporary plugins directory: %v", err)
	}

	repoPath := filepath.Join(tempDir, repoName)
	logrus.Infof("Cloning plugin from %s to %s", repoURL, repoPath)
	err := pr.gitPluginProvider.PullGitHubPlugin(repoPath, repoURL)
	if err != nil {
		return "", fmt.Errorf("An error occurred pulling plugin from GitHub:\n%v", err.Error())
	}

	return repoPath, nil
}

func createVirtualEnv(venvPath string) error {
	cmd := exec.Command("python3", "-m", "venv", venvPath)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to create virtual environment: %v\nOutput: %s", err, output)
	}
	return nil
}

func installDependencies(venvPath, requirementsPath string) error {
	cmd := exec.Command(filepath.Join(venvPath, "bin", "pip"), "install", "-r", requirementsPath)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to install dependencies: %v\nOutput: %s", err, output)
	}
	return nil
}
