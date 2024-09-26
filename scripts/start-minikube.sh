#!/bin/zsh

kubectl config set-context minikube

minikube start --driver=docker

minikube addons enable ingress

minikube addons enable metrics-server

istioctl install --set profile=minimal --set meshConfig.accessLogFile=/dev/stdout -y

kubectl apply -f https://github.com/kubernetes-sigs/gateway-api/releases/download/v1.1.0/standard-install.yaml

kubectl apply -f https://raw.githubusercontent.com/istio/istio/release-1.10/samples/addons/prometheus.yaml
kubectl apply -f https://raw.githubusercontent.com/istio/istio/release-1.10/samples/addons/grafana.yaml
kubectl apply -f https://raw.githubusercontent.com/istio/istio/release-1.10/samples/addons/jaeger.yaml
kubectl apply -f https://raw.githubusercontent.com/istio/istio/release-1.10/samples/addons/kiali.yaml
kubectl rollout status deployment/kiali -n istio-system

minikube dashboard
