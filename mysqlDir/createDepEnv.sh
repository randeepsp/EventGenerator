#!/bin/bash

echo "building mysql image"
docker  build -t rmysql .
echo " now createing k8s namespace env"
kubectl apply -f knamespace.yaml
echo "now creating mysql env"
kubectl apply -f mysql-dep.yaml
echo "completed deploying mysql pod"
