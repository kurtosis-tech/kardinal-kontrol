#!/bin/bash

# Check for required arguments
if [ "$#" -ne 1 ]; then
  echo "Usage: $0 <namespace>"
  exit 1
fi

NAMESPACE=$1

# Function to check if port-forwarding is needed and establish it
setup_port_forwarding() {
    echo "Setting up port forwarding for Kiali in namespace istio-system..."
    kubectl port-forward svc/kiali -n istio-system 20001:20001 &
    PF_PID=$!
    echo "Port forwarding established with PID: $PF_PID"
    sleep 5  # Wait for the port-forward to be established
}

# Function to fetch the graph data
fetch_graph_data() {
    echo "Fetching graph data for namespace $NAMESPACE..."
    curl "http://localhost:20001/kiali/api/namespaces/graph?duration=60s&graphType=versionedApp&includeIdleEdges=false&injectServiceNodes=true&boxBy=cluster,namespace&appenders=deadNode,istio,serviceEntry,meshCheck,workloadEntry,health&rateGrpc=requests&rateHttp=requests&rateTcp=sent&namespaces=$NAMESPACE" -o "graph.json"
    if [ $? -eq 0 ]; then
        echo "Graph data successfully fetched and saved to graph.json"
    else
        echo "Failed to fetch graph data"
    fi
}

# Main execution function
main() {
    setup_port_forwarding
    fetch_graph_data
    kill $PF_PID
    echo "Port forwarding stopped."
}

# Execute the main function
main
