package plugins

var MockGitHub = map[string]map[string]string{
	// Simple Plugin
	"https://github.com/fake-org/kardinal-simple-plugin-example.git": {
		"main.py": `import copy

REPLACED = "the-text-has-been-replaced"


def create_flow(service_specs: list, deployment_specs: list, flow_uuid, text_to_replace):
    modified_deployment_specs = []

    for deployment_spec in deployment_specs:
        modified_deployment_spec = copy.deepcopy(deployment_spec)
        modified_deployment_spec['template']['metadata']['labels']['app'] = modified_deployment_spec['template']['metadata']['labels']['app'].replace(text_to_replace, REPLACED)
        modified_deployment_spec['selector']['matchLabels']['app'] = modified_deployment_spec['selector']['matchLabels']['app'].replace(text_to_replace, REPLACED)
        modified_deployment_spec['template']['spec']['containers'][0]['name'] = modified_deployment_spec['template']['spec']['containers'][0]['name'].replace(text_to_replace, REPLACED)
        
        modified_deployment_specs.append(modified_deployment_spec)

    config_map = {
        "original_text": text_to_replace
    }
  
    return {
        "deployment_specs": modified_deployment_specs,
        "config_map": config_map
    }


def delete_flow(config_map, flow_uuid):
    print(config_map["original_text"])
`,
	},
	// Complex Plugin
	"https://github.com/fake-org/kardinal-slightly-more-complex-plugin-example.git": {
		"main.py": `import copy
import requests


def create_flow(service_specs: list, deployment_specs: list, flow_uuid):
    response = requests.get("https://4.ident.me")
    if response.status_code != 200:
        raise Exception("An unexpected error occurred")

    ip_address = response.text.strip()

    modified_deployment_specs = []

    for deployment_spec in deployment_specs:
        modified_deployment_spec = copy.deepcopy(deployment_spec)
        # Replace the IP address in the environment variable
        for container in modified_deployment_spec['template']['spec']['containers']:
            for env in container['env']:
                if env['name'] == 'REDIS':
                    env['value'] = ip_address

        modified_deployment_specs.append(modified_deployment_spec)


    config_map = {
        "original_value": "ip_addr"
    }

    return {
        "deployment_specs": modified_deployment_specs,
        "config_map": config_map
    }


def delete_flow(config_map, flow_uuid):
    # In this complex plugin, we don't need to do anything for deletion
    return None`,
		"requirements.txt": "requests",
	},
	// Identity Plugin
	"https://github.com/fake-org/kardinal-identity-plugin-example.git": {
		"main.py": `def create_flow(service_specs, deployment_specs, flow_uuid):
    return {
        "deployment_specs": deployment_specs,
        "config_map": {}
    }

def delete_flow(config_map, flow_uuid):
    return None
`,
	},
	// Redis sidecar plugin
	"https://github.com/fake-org/kardinal-redis-sidecar-plugin-example.git": {
		"main.py": `import copy


def create_flow(service_specs, deployment_specs, flow_uuid):

    modified_deployment_specs = []

    for deployment_spec in deployment_specs:
        modified_deployment_spec = copy.deepcopy(deployment_spec)
        modified_deployment_spec['template']['spec']['containers'][0]["image"] = "kurtosistech/redis-proxy-overlay:latest"
    
        modified_deployment_specs.append(modified_deployment_spec)
    
    return {
        "deployment_specs": modified_deployment_specs,
        "config_map": {}
    }

def delete_flow(config_map, flow_uuid):
    pass

`,
	},
}
