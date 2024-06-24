# Kardinal

## Developing instructions

1. Enter the dev shell and start the local cluster:

```bash
nix develop
```

2. You're also likely to use a local k8s, in this case minikube is available to use:

```bash
kubectl config set-context minikube
minikube start --driver=docker --cpus=10 --memory 8192 --disk-size 32g
minikube addons enable ingress
minikube addons enable metrics-server
istioctl install --set profile=demo -y
minikube dashboard
```

On a second terminal, start the tunnel:

```bash
minikube tunnel
```

## Deploying Kontrol to local cluster

You can use tilt deploy and keeping the image hot-reloading:

```bash
tilt up
```

Or manually build it:

```bash
# First set the docker context to minikube
eval $(minikube docker-env)
docker load < $(nix build ./#kardinal-manager-container --no-link --print-out-paths)
kubectl apply -f kontrol-service/deployment
```

## Deploying Redis Overlay Service to local cluster

Building and loading image into minikube:

```bash
# First set the docker context to minikube
eval $(minikube docker-env)
docker load < $(nix build ./#redis-proxy-overlay-container --no-link --print-out-paths)
```

To build and run the service directly:

```bash
nix run ./#redis-proxy-overlay

```

## Deploying Kloud Kontrol service to local cluster

Building and loading image into minikube:

```bash
# First set the docker context to minikube
eval $(minikube docker-env)
docker load < $(nix build ./#kloud-kontrol-container --no-link --print-out-paths)
```

To build and run the service directly:

```bash
nix run ./#kloud-kontrol

```

## Running Kardinal CLI

To build and run the service directly:

````bash
nix run ./#kardinal-cli
```


### Regenerate REST API Bindings

You can either:

1. Press the green play button in `kontrol-service/kardinal-manager/api/http_rest/generate.go` in Goland
   <img src="./.github/readme-static-files/goland-generate-rest-bindings.png"/>
2. Or execute the following go from the repository root

```bash
go generate ./kontrol-service/kardinal-manager/api/http_rest/generate.go
````

### Regenerate gomod2nix.toml

You will need to do this every time the `go.mod` file is edited

```bash
# inside the kontrol-service directory
nix develop
gomod2nix generate
```
