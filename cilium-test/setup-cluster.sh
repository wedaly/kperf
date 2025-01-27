#!/usr/bin/env bash

# Assume that kubectl is already configured with access to the test cluster.
# That cluster should NOT have cilium installed -- we'll create the Cilium custom resources ourselves
# and don't want any of the cilium components modifying these.

# First, load the Cilium CRDs. Arbitrarily use Cilium 1.16.
kubectl apply -f https://raw.githubusercontent.com/cilium/cilium/refs/tags/v1.16.6/pkg/k8s/apis/cilium.io/client/crds/v2/ciliumendpoints.yaml
kubectl apply -f https://raw.githubusercontent.com/cilium/cilium/refs/tags/v1.16.6/pkg/k8s/apis/cilium.io/client/crds/v2/ciliumidentities.yaml

# Next, create CiliumEndpoint and CiliumIdentity custom resources.
# These are the Cilium resources that scale with the number of pods in a cluster
# and generate the most load on apiserver.
# The contents of these resources don't particularly matter for this test; we're mostly
# interested in the total count and size of the objects.