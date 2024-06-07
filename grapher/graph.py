import json
import yaml

# Load JSON from a file
with open('graph.json', 'r') as file:
    data = json.load(file)

# Extract relevant data
services = {}
for node in data['elements']['nodes']:
    if 'app' in node['data']:
        service_name = node['data'].get('app')
        version = node['data'].get('version', 'latest')  # Default to 'latest' if version is not specified
        node_id = node['data']['id']
        services[node_id] = {
            'name': service_name,
            'version': version,
            'talks_to': []
        }

# Map the edges to show communications
for edge in data['elements']['edges']:
    source_id = edge['data']['source']
    target_id = edge['data']['target']
    if source_id in services and target_id in services:
        source_key = f"{services[source_id]['name']}_{services[source_id]['version']}"
        target_key = f"{services[target_id]['name']}_{services[target_id]['version']}"
        services[source_id]['talks_to'].append(target_key)

# Create a simplified dictionary to convert to YAML
output = {f"{service['name']}_{service['version']}": {
                'version': service['version'], 
                'talks_to': service['talks_to'],
                'name': service['name'],
            } for service in services.values()}

# Convert to YAML
yaml_data = yaml.dump(output, allow_unicode=True, sort_keys=False)

# Print the YAML data
print(yaml_data)
