#!/bin/zsh

kubectl config set-context minikube

minikube start --driver=docker

minikube addons enable ingress

minikube addons enable metrics-server

istioctl install --set profile=minimal --set meshConfig.accessLogFile=/dev/stdout -y

minikube dashboard
