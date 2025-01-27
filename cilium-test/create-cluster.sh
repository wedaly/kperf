#!/usr/bin/env bash

SUBSCRIPTION=""
RESOURCE_GROUP=""
CLUSTER_NAME=""

az account set -s "$SUBSCRIPTION"
az aks create -g "$RESOURCE_GROUP" -n "$CLUSTER_NAME" \
    --network-plugin azure \
    --network-plugin-mode overlay \
    --network-dataplane azure

az aks nodepool add -g "$RESOURCE_GROUP" --cluster-name "$CLUSTER_NAME" -n kperf \
    --node-vm-size Standard_D16s_v3 \
    --node-count 6 

# please write Azure CLI commands to create an AKS cluster with:
# networkPlugin=azure
# networkPluginMode=overlay
# networkDataplane=azure


