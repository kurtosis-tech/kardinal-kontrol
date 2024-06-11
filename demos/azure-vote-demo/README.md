# Azure Voting Application + Redis Proxy Overlay

We added a Redis proxy in between the Azure vote frontend and backend. The proxy supports get, set and incrby commands with an in-memory cache. The missed gets are proxied to the Azure vote backend and their result stored in the in-memory cache.

## Deploying

1. You're also likely to use a local k8s, in this case minikube is available to use:

```bash
kubectl config set-context minikube
minikube start --driver=docker --cpus=10 --memory 8192 --disk-size 32g
minikube addons enable ingress
minikube addons enable metrics-server
minikube dashboard
```

2. You will need to install Istio and its addons in the local cluster:

```bash
# Install Istio in the local cluster with the demo profile
istioctl install --set profile=demo -y

# Install Kiali and the other Addons
kubectl apply -f https://raw.githubusercontent.com/istio/istio/release-1.10/samples/addons/prometheus.yaml
kubectl apply -f https://raw.githubusercontent.com/istio/istio/release-1.10/samples/addons/grafana.yaml
kubectl apply -f https://raw.githubusercontent.com/istio/istio/release-1.10/samples/addons/jaeger.yaml
kubectl apply -f https://raw.githubusercontent.com/istio/istio/release-1.10/samples/addons/kiali.yaml
kubectl rollout status deployment/kiali -n istio-system

# Access into the Kiali dashboard
istioctl dashboard kiali
```

3. Build the Redis proxy overlay and load it into Minikube.

```bash
eval $(minikube docker-env)
docker load < $(nix build ./#redis-proxy-overlay-container --no-link --print-out-paths)
minikube image build -t voting-app-ui -f ./Dockerfile ./demos/azure-vote-demo/voting-app-ui/
```

4. Deploy the Azure voting app and Redis proxy overlay.

```bash
kubectl create namespace voting-app
kubectl label namespace voting-app istio-injection=enabled
kubectl apply -n voting-app -f demos/azure-vote-demo/prod-only-demo.yaml
```

5. On another terminal, start the tunnel:

```bash
minikube tunnel
```
