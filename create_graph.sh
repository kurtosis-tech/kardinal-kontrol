#!/bin/bash

# Check for required arguments
if [ "$#" -ne 1 ]; then
  echo "Usage: $0 <namespace>"
  exit 1
fi

NAMESPACE=$1

# Function to check if Istio is already installed
check_istio_installed() {
  if istioctl version --remote=false > /dev/null 2>&1; then
    echo "Istio is already installed."
  else
    echo "Istio is not installed. Installing Istio..."
    curl -L https://istio.io/downloadIstio | sh -
    cd istio-* || exit
    export PATH=$PWD/bin:$PATH
    istioctl install --set profile=demo -y
  fi
}

# Function to check if Kiali is installed
check_kiali_installed() {
  if kubectl get deployment kiali -n istio-system > /dev/null 2>&1; then
    echo "Kiali is already installed."
  else
    echo "Kiali is not installed. Installing Kiali..."
    kubectl apply -f samples/addons
    kubectl rollout status deployment/kiali -n istio-system
  fi
}

# Install Istio if it is not installed
check_istio_installed

# Install Kiali if it is not installed
check_kiali_installed

# Generate the YAML file for the namespace's graph
echo "Generating YAML for service graph in namespace $NAMESPACE..."
kubectl port-forward svc/kiali -n istio-system 20001:20001 &
PF_PID=$!

sleep 5  # Wait for the port-forward to establish

# Fetch the graph data
curl -s "http://localhost:20001/kiali/api/namespaces/$NAMESPACE/graph?type=app&duration=1h&graphType=versionedApp" -o "${NAMESPACE}_graph.yaml"

kill $PF_PID
echo "Graph YAML saved as ${NAMESPACE}_graph.yaml"

