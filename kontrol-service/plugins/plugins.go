package plugins

import (
	"encoding/json"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	appv1 "k8s.io/api/apps/v1"
)

type PluginRunner struct {
	memory map[string]string
}

func NewPluginRunner() *PluginRunner {
	return &PluginRunner{
		memory: make(map[string]string),
	}
}

func (pr *PluginRunner) CreateFlow(pluginUrl string, serviceSpec corev1.ServiceSpec, deploymentSpec appv1.DeploymentSpec, flowUuid string, arguments map[string]string) (appv1.DeploymentSpec, string, error) {
	repoPath, err := getOrCloneRepo(pluginUrl)
	if err != nil {
		return appv1.DeploymentSpec{}, "", fmt.Errorf("failed to get or clone repository: %v", err)
	}

	serviceSpecJSON, err := json.Marshal(serviceSpec)
	if err != nil {
		return appv1.DeploymentSpec{}, "", fmt.Errorf("failed to marshal service spec: %v", err)
	}

	deploymentSpecJSON, err := json.Marshal(deploymentSpec)
	if err != nil {
		return appv1.DeploymentSpec{}, "", fmt.Errorf("failed to marshal deployment spec: %v", err)
	}

	result, err := runPythonCreateFlow(repoPath, string(serviceSpecJSON), string(deploymentSpecJSON), flowUuid, arguments)
	if err != nil {
		return appv1.DeploymentSpec{}, "", err
	}

	var resultMap map[string]json.RawMessage
	err = json.Unmarshal([]byte(result), &resultMap)
	if err != nil {
		return appv1.DeploymentSpec{}, "", fmt.Errorf("failed to parse result: %v", err)
	}

	var newDeploymentSpec appv1.DeploymentSpec
	err = json.Unmarshal(resultMap["deployment_spec"], &newDeploymentSpec)
	if err != nil {
		return appv1.DeploymentSpec{}, "", fmt.Errorf("failed to unmarshal deployment spec: %v", err)
	}

	configMapJSON := resultMap["config_map"]
	var configMap map[string]interface{}
	err = json.Unmarshal(configMapJSON, &configMap)
	if err != nil {
		return appv1.DeploymentSpec{}, "", fmt.Errorf("invalid config map: %v", err)
	}

	configMapString, err := json.Marshal(configMap)
	if err != nil {
		return appv1.DeploymentSpec{}, "", fmt.Errorf("failed to re-marshal config map: %v", err)
	}

	pr.memory[flowUuid] = string(configMapString)

	return newDeploymentSpec, string(configMapString), nil
}

func (pr *PluginRunner) DeleteFlow(pluginUrl, flowUuid string, arguments map[string]string) error {
	repoPath, err := getOrCloneRepo(pluginUrl)
	if err != nil {
		return fmt.Errorf("failed to get or clone repository: %v", err)
	}

	configMap, ok := pr.memory[flowUuid]
	if !ok {
		return fmt.Errorf("no config map found for flow UUID: %s", flowUuid)
	}

	_, err = runPythonDeleteFlow(repoPath, configMap, flowUuid, arguments)
	if err != nil {
		return err
	}

	delete(pr.memory, flowUuid)
	return nil
}

// runPythonCreateFlow this runs the create_flow plugin function
// TODO (gm)
// -- harden this - currently this captures the result value of creat_flow via a print statement which is fragile
// a user could have a print statements anywhere else. a way to fix this could be to write to a file or some use other form of IPC
func runPythonCreateFlow(repoPath, serviceSpecJSON, deploymentSpecJSON, flowUuid string, arguments map[string]string) (string, error) {
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

	argsJSON, err := json.Marshal(arguments)
	if err != nil {
		return "", fmt.Errorf("failed to marshal arguments: %v", err)
	}
	argJsonStr := string(argsJSON)

	tempScript := fmt.Sprintf(`
import sys
import json
import inspect
sys.path.append("%s")
import main

service_spec = json.loads('''%s''')
deployment_spec = json.loads('''%s''')
flow_uuid = %q
args = json.loads('''%s''')

# Get the signature of the create_flow function
sig = inspect.signature(main.create_flow)

# Prepare kwargs based on the function signature
kwargs = {}
for param in sig.parameters.values():
    if param.name == 'service_spec':
        kwargs['service_spec'] = service_spec
    elif param.name == 'deployment_spec':
        kwargs['deployment_spec'] = deployment_spec
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
print(json.dumps(result))
`, repoPath, serviceSpecJSON, deploymentSpecJSON, flowUuid, argJsonStr)

	return executePythonScript(venvPath, repoPath, tempScript)
}

func runPythonDeleteFlow(repoPath, configMap, flowUuid string, arguments map[string]string) (string, error) {
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

	argsJSON, err := json.Marshal(arguments)
	if err != nil {
		return "", fmt.Errorf("failed to marshal arguments: %v", err)
	}

	tempScript := fmt.Sprintf(`
import sys
import json
sys.path.append("%s")
import main
import inspect

config_map = %s
flow_uuid = %q
args = json.loads('''%s''')
sig = inspect.signature(main.delete_flow)

kwargs = {}
for param in sig.parameters.values():
    if param.name == 'flow_uuid':
        kwargs['flow_uuid'] = flow_uuid
    elif param.name == 'config_map':
        kwargs['config_map'] = config_map
    elif param.name in args:
        kwargs[param.name] = args[param.name]
    elif param.default is not param.empty:
        kwargs[param.name] = param.default
    else:
        print(f"Warning: Required parameter {param.name} not provided")

# Call create_flow with the prepared kwargs
result = main.delete_flow(**kwargs)
print(json.dumps(result))
`, repoPath, configMap, flowUuid, argsJSON)

	return executePythonScript(venvPath, repoPath, tempScript)
}

func executePythonScript(venvPath, repoPath, scriptContent string) (string, error) {
	tempFile, err := os.CreateTemp("", "temp_script_*.py")
	if err != nil {
		return "", fmt.Errorf("failed to create temporary script: %v", err)
	}
	defer os.Remove(tempFile.Name())

	if _, err := tempFile.WriteString(scriptContent); err != nil {
		return "", fmt.Errorf("failed to write temporary script: %v", err)
	}
	tempFile.Close()

	cmd := exec.Command(filepath.Join(venvPath, "bin", "python"), tempFile.Name())
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to run Python script: %v\nOutput: %s", err, output)
	}

	return strings.TrimSpace(string(output)), nil
}

func getOrCloneRepo(repoURL string) (string, error) {
	if !strings.HasPrefix(repoURL, "https://") {
		repoURL = "https://" + repoURL
	}

	parts := strings.Split(repoURL, "/")
	repoName := parts[len(parts)-1]

	tempDir := filepath.Join(os.TempDir(), "go-python-plugins")
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create temporary plugins directory: %v", err)
	}

	repoPath := filepath.Join(tempDir, repoName)

	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		cmd := exec.Command("git", "clone", repoURL, repoPath)
		if output, err := cmd.CombinedOutput(); err != nil {
			return "", fmt.Errorf("git clone failed: %v\nOutput: %s", err, output)
		}
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
