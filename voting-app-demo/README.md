# Azure Voting Application + Redis Proxy Overlay

We added a Redis proxy in between the Azure vote frontend and backend. The proxy supports get, set and incrby commands with an in-memory cache. The missed gets are proxied to the Azure vote backend and their result stored in the in-memory cache.

## Deploying

1. You're also likely to use a local k8s, in this case minikube is available to use:

```bash
nix develop

kubectl config set-context minikube
minikube start --driver=docker --cpus=10 --memory 8192 --disk-size 32g
minikube addons enable ingress
minikube addons enable metrics-server
minikube dashboard
```

2. You will need to install Istio and its addons in the local cluster:

```bash
nix develop

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
nix develop
eval $(minikube docker-env)
docker load < $(nix build ./#redis-proxy-overlay-container --no-link --print-out-paths)
minikube image build -t voting-app-ui -f ./Dockerfile ./voting-app-demo/voting-app-ui/
```

4. Deploy the Azure voting app and Redis proxy overlay.

```bash
nix develop
kubectl create namespace voting-app
kubectl label namespace voting-app istio-injection=enabled
kubectl apply -n voting-app -f ./voting-app-demo/manifests/prod-only-demo.yaml
minikube tunnel
```

Add the hots for test configuration in the host file

```bash
sudo nano /private/etc/hosts
```

And include these lines at the end and save the host file

```bash
127.0.0.1	voting-app.localhost
127.0.0.1	dev.voting-app.localhost
```

## Demo

After deploying the application, you can access the Azure voting app at [http://voting-app.local](http://voting-app.local). And
can also start some artificial load with the following command (`nix develop` will make them available in the shell):

```bash
load-generator
```

After some time, you can access the Kiali dashboard to see the traffic flow between the services in the production mode. Now
you enter the dev mode and start to test with the Redis proxy overlay.

```bash
kardinal create-dev-flow voting-app test
```

Youn can now access the dev path at [http://dev.voting-app.local](http://dev.voting-app.local) and the Kiali dashboard will reflect the new traffic flow.
Use the following command to reset the state (replace the pod) on Redis proxy overlay:

```bash
kardinal reset-dev-flow voting-app
```

And finally, you can delete the dev path with the following command:

```bash
kardinal delete-dev-flow voting-app
```