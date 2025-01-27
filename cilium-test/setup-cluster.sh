#!/usr/bin/env bash

set -ex

CILIUM_VERSION="v1.16.6"
NUM_CILIUM_ENDPOINTS=10
NUM_CILIUM_IDENTITIES=10

# Assume that kubectl is already configured with access to the test cluster.
# That cluster should NOT have cilium installed -- we'll create the Cilium custom resources ourselves
# and don't want any of the cilium components modifying these.

# Some helper functions
random_label_val() {
  echo "\"$(head /dev/urandom | tr -dc A-Za-z0-9 | head -c64)\""
}

# First, load the Cilium CRDs. Arbitrarily use Cilium 1.16.
kubectl apply -f "https://raw.githubusercontent.com/cilium/cilium/refs/tags/$CILIUM_VERSION/pkg/k8s/apis/cilium.io/client/crds/v2/ciliumendpoints.yaml"
kubectl apply -f "https://raw.githubusercontent.com/cilium/cilium/refs/tags/$CILIUM_VERSION/pkg/k8s/apis/cilium.io/client/crds/v2/ciliumidentities.yaml"

# Next, create CiliumEndpoint and CiliumIdentity custom resources.
# These are the Cilium resources that scale with the number of pods in a cluster
# and generate the most load on apiserver.
# The contents of these resources don't particularly matter for this test; we're mostly
# interested in the total count and size of the objects.
initial_num_cid="$(kubectl get ciliumidentities --no-headers | wc -l || "0")"
for i in $(seq $initial_num_cid $NUM_CILIUM_IDENTITIES); do
cid_name=$(printf "%06d" $i)
cat <<EOF | kubectl apply -f -
apiVersion: cilium.io/v2
kind: CiliumIdentity
metadata:
  name: "$cid_name"
security-labels:
  random01: $(random_label_val)
  random02: $(random_label_val)
  random03: $(random_label_val)
  random04: $(random_label_val)
  random05: $(random_label_val)
  random06: $(random_label_val)
  random07: $(random_label_val)
  random08: $(random_label_val)
  random09: $(random_label_val)
EOF
done

initial_num_cep="$(kubectl get ciliumendpoints --no-headers | wc -l)"
num_cep_to_create=$((NUM_CILIUM_ENDPOINTS - initial_num_cep))
for i in $(seq 1 $num_cep_to_create); do
cat <<EOF | kubectl apply -f -
    
EOF
done