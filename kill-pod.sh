#!/bin/sh

NAMESPACE=hive-operator-system
NAME=$(kubectl get pods -n $NAMESPACE -o name)
sudo kubectl delete $NAME  -n $NAMESPACE
