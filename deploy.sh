#!/bin/bash
set -e
kubectl delete pods --all
make images
docker push gcr.io/priya-wadhwa/executor:latest
kubectl create -f pod.yaml