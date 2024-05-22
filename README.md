# Kardinal Progressive Delivery Demo

This repo contains the [Argo Rollouts](https://github.com/argoproj/argo-rollouts) demo application source code and examples. It demonstrates the
various deployment strategies and progressive delivery features of Argo Rollouts.

Before running an example:

1. Enter the dev shell and start the local cluster:

```bash
nix develop
minikube start --driver=docker --cpus=8 --memory 8192 --disk-size 32g
kubectl config set-context minikube
# You can also start the minikube dashboard in a separate terminal
minikube dashboard
```

2. Kardinal demo

```bash
kubectl create namespace kardinal-demo
kubectl apply -n kardinal-demo -f https://github.com/argoproj/argo-rollouts/releases/latest/download/install.yaml
kubectl apply -n kardinal-demo -f kardinal-demo
kubectl port-forward -n kardinal-demo service/frontend 8080:80
```

3. Google microservices demo (optional)

```bash
kubectl create namespace ms-demo
kubectl apply -n ms-demo -f microservices-demo
# or directly from the Github repo
# kubectl apply -n ms-demo -f https://raw.githubusercontent.com/GoogleCloudPlatform/microservices-demo/main/release/kubernetes-manifests.yaml
kubectl port-forward -n ms-demo deployment/frontend 8080:8080
```

4.  Argo B/G Demo (Optional)

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

```bash
kubectl argo rollouts -n argo-demo dashboard
```
