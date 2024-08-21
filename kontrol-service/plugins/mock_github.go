package plugins

var MockGitHub = map[string]map[string]string{
	// Simple Plugin
	"https://github.com/h4ck3rk3y/a-test-plugin.git": {
		"main.py": `REPLACED = "the-text-has-been-replaced"

def create_flow(service_spec, deployment_spec, flow_uuid, text_to_replace):
    deployment_spec['template']['metadata']['labels']['app'] = deployment_spec['template']['metadata']['labels']['app'].replace(text_to_replace, REPLACED)
    deployment_spec['selector']['matchLabels']['app'] = deployment_spec['selector']['matchLabels']['app'].replace(text_to_replace, REPLACED)
    deployment_spec['template']['spec']['containers'][0]['name'] = deployment_spec['template']['spec']['containers'][0]['name'].replace(text_to_replace, REPLACED)
    
    config_map = {
        "original_text": text_to_replace
    }
    
    return {
        "deployment_spec": deployment_spec,
        "config_map": config_map
    }

def delete_flow(config_map, flow_uuid):
    print(config_map["original_text"])
`,
	},
	// Complex Plugin
	"https://github.com/h4ck3rk3y/slightly-more-complex-plugin.git": {
		"main.py": `import json
import requests

def create_flow(service_spec, deployment_spec, flow_uuid):
    response = requests.get("https://ident.me")
    if response.status_code != 200:
        raise Exception("An unexpected error occurred")
    
    ip_address = response.text.strip()
    
    # Replace the IP address in the environment variable
    for container in deployment_spec['template']['spec']['containers']:
        for env in container['env']:
            if env['name'] == 'REDIS':
                env['value'] = ip_address
    
    config_map = {
        "original_value": "ip_addr"
    }
    
    return {
        "deployment_spec": deployment_spec,
        "config_map": config_map
    }

def delete_flow(config_map, flow_uuid):
    # In this complex plugin, we don't need to do anything for deletion
    return None`,
		"requirements.txt": "requests",
	},
	// Identity Plugin
	"https://github.com/h4ck3rk3y/identity-plugin.git": {
		"main.py": `def create_flow(service_spec, deployment_spec, flow_uuid):
    return {
        "deployment_spec": deployment_spec,
        "config_map": {}
    }

def delete_flow(config_map, flow_uuid):
    return None
`,
	},
	// Redis sidecar plugin
	"https://github.com/h4ck3rk3y/redis-sidecar-plugin.git": {
		"main.py": `def create_flow(service_spec, deployment_spec, flow_uuid):
    deployment_spec['template']['spec']['containers'][0]["image"] = "kurtosistech/redis-proxy-overlay:latest"
    return {
        "deployment_spec": deployment_spec,
        "config_map": {}
    }

def delete_flow(config_map, flow_uuid):
    pass
`,
	},
}
