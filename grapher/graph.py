import json
import yaml

# Load JSON from a file (assuming you've saved the JSON data into a file called 'graph.json')
with open('graph.json', 'r') as file:
    data = json.load(file)

# Extract relevant data
services = {}
for node in data['elements']['nodes']:
    if 'app' in node['data']:
        service_name = node['data'].get('app')
        version = node['data'].get('version', 'latest')  # Default to 'latest' if version is not specified
        services[node['data']['id']] = {'name': service_name, 'version': version, 'talks_to': []}

# Map the edges to show communications
for edge in data['elements']['edges']:
    source_id = edge['data']['source']
    target_id = edge['data']['target']
    if source_id in services and target_id in services:
        services[source_id]['talks_to'].append(services[target_id]['name'])

# Create a simplified dictionary to convert to YAML
output = {service['name']: {'version': service['version'], 'talks_to': service['talks_to']} for service in services.values()}

# Convert to YAML
yaml_data = yaml.dump(output, allow_unicode=True)

# Print the YAML data
print(yaml_data)
