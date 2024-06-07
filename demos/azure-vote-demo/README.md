# Azure Voting Application + Redis Proxy Overlay

We added a Redis proxy in between the Azure vote frontend and backend. The proxy supports get, set and incrby commands with an in-memory cache. The missed gets are proxied to the Azure vote backend and their result stored in the in-memory cache.

## Deploying

Build the Redis proxy overlay and load it into Minikube.

```bash
eval $(minikube docker-env)
docker load < $(nix build ./#redis-proxy-overlay-container --no-link --print-out-paths)
```

Deploy the Azure voting app and Redis proxy overlay.

```bash
kubectl create namespace voting-app
kubectl label namespace voting-app istio-injection=enabled
kubectl apply -n voting-app -f demos/azure-vote-demo/prod-only-demo.yaml
```
