[![Private Registry](https://img.shields.io/badge/AWS_ECR-images-important.svg?logo=awsfargate)](https://us-east-1.console.aws.amazon.com/ecr/private-registry/repositories?region=us-east-1)
[![Deployed Cluster](https://img.shields.io/badge/AWS_EKS-cluster-important.svg?logo=awsfargate)](https://us-east-1.console.aws.amazon.com/eks/home?region=us-east-1#/clusters/kardinal-kluster)

# Kardinal Kontrol

## Disclaimer: This project is no longer maintained.

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

## Publishing multi-arch images

To publish multi-arch images, you can use the following command:

```bash
$(nix build .#publish-<SERVICE_NAME>-container --no-link --print-out-paths)/bin/push

# For instance, to publish the kontrol-service image:
$(nix build .#publish-kontrol-service-container --no-link --print-out-paths)/bin/push
```

## Deploying Kontrol service to local cluster

Building and loading image into minikube:

```bash
# First set the docker context to minikube
eval $(minikube docker-env)
docker load < $(nix build ./#kontrol-service-container --no-link --print-out-paths)
```

To build and run the service directly in dev mode (a locally running Postgres DB is required):

```bash
DB_HOSTNAME=localhost DB_USERNAME=postgres DB_NAME=kardinal DB_PORT=5432 DB_PASSWORD=<database password> nix run ./#kontrol-service -- -dev-mode
```

### Regenerate gomod2nix.toml

You will need to do this every time a `go.mod` file is edited

```bash
nix develop
gomod2nix generate
```

