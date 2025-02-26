#!/bin/sh

kubectl config delete-cluster hive
kind delete cluster --name hive

