# Kardinal

## Demos

This repo contains the [Argo Rollouts](https://github.com/argoproj/argo-rollouts) demo application source code and examples. It demonstrates the
various deployment strategies and progressive delivery features of Argo Rollouts.

Before running an example:

1. Enter the dev shell and start the local cluster:

```bash
nix develop
kubectl config set-context minikube
minikube start --driver=docker --cpus=10 --memory 8192 --disk-size 32g
minikube addons enable ingress
minikube addons enable metrics-server
minikube dashboard
```

On another terminal, start the tunnel:

```bash
minikube tunnel
```

And if using the Argo Rollout, start the dashboard (optional):

```bash
kubectl apply -n <namespace> -f rbac-rollout-dashboard.yaml
kubectl port-forward -n <namespace> services/argo-rollouts-dashboard 3100:3100
```

<details>
  <summary>Kardinal demo</summary>

```bash
cd demos
kubectl create namespace kardinal-demo
kubectl apply -n kardinal-demo -f https://github.com/argoproj/argo-rollouts/releases/latest/download/install.yaml
kubectl apply -n kardinal-demo -f kardinal-demo
kubectl port-forward -n kardinal-demo service/frontend 8080:80
```

Start to rollout changes to the frontend service:

```bash
kubectl argo rollouts -n kardinal-demo set image frontend "*=lostbean/microservice-frontend:banner1"
kubectl argo rollouts -n kardinal-demo set image frontend "*=lostbean/microservice-frontend:banner2"
```

</details>

<details>
  <summary>Google microservices demo (optional)</summary>

```bash
kubectl create namespace ms-demo
kubectl apply -n ms-demo -f microservices-demo
# or directly from the Github repo
# kubectl apply -n ms-demo -f https://raw.githubusercontent.com/GoogleCloudPlatform/microservices-demo/main/release/kubernetes-manifests.yaml
kubectl port-forward -n ms-demo deployment/frontend 8080:8080
```

</details>

<details>
  <summary>Argo B/G Demo (Optional)</summary>

```bash
kubectl create namespace argo-demo
kubectl apply -n argo-demo -f https://github.com/argoproj/argo-rollouts/releases/latest/download/install.yaml
kubectl apply -n argo-demo -f argo-demo
```

```bash
minikube service -n argo-demo bluegreen-demo --url
minikube service -n argo-demo bluegreen-demo-preview --url
```

Watch the rollout or experiment using the argo rollouts kubectl plugin:

```bash
kubectl argo rollouts -n argo-demo get rollout bluegreen-demo --watch
```

For rollouts, trigger an update by setting the image of a new color to run:

```bash
kubectl argo rollouts -n argo-demo set image bluegreen-demo "*=argoproj/rollouts-demo:yellow"
```

</details>
