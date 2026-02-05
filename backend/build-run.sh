#!/bin/sh

echo "bulid docker"
make docker-build
sleep 1
echo "上线"
kubectl replace --force -f deploy/deployment.yaml -n k8s-manager
sleep 1
kubectl get pod  -n k8s-manager
