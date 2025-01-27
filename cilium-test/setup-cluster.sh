#!/usr/bin/env bash

set -e

CILIUM_VERSION="v1.16.6"
NUM_CILIUM_ENDPOINTS=50000 # 5k nodes x 100 pods per node
NUM_CILIUM_IDENTITIES=5000 # based on outages we've seen with CID spikes
BATCH_SIZE=1000

# Assume that kubectl is already configured with access to the test cluster.
# That cluster should NOT have cilium installed -- we'll create the Cilium custom resources ourselves
# and don't want any of the cilium components modifying these.

# First, load the Cilium CRDs. Arbitrarily use Cilium 1.16.
echo "Loading Cilium CRDs"
kubectl apply -f "https://raw.githubusercontent.com/cilium/cilium/refs/tags/$CILIUM_VERSION/pkg/k8s/apis/cilium.io/client/crds/v2/ciliumendpoints.yaml"
kubectl apply -f "https://raw.githubusercontent.com/cilium/cilium/refs/tags/$CILIUM_VERSION/pkg/k8s/apis/cilium.io/client/crds/v2/ciliumidentities.yaml"

# Create CiliumIdentity custom resources in batches
initial_num_cid="$(kubectl get ciliumidentities --no-headers | wc -l || "0")"
for ((i=initial_num_cid; i<=$NUM_CILIUM_IDENTITIES; i+=BATCH_SIZE)); do
  echo "Creating CiliumIdentity batch $((i/BATCH_SIZE + 1))"
  batch=""
  for ((j=i; j<i+BATCH_SIZE && j<=$NUM_CILIUM_IDENTITIES; j++)); do
    cid_name=$(printf "%06d" $j)
    batch+=$(cat <<EOF

apiVersion: cilium.io/v2
kind: CiliumIdentity
metadata:
  name: "$cid_name"
security-labels:
  k8s:io.cilium.k8s.namespace.labels.addonmanager.kubernetes.io/mode: Reconcile
  k8s:io.cilium.k8s.namespace.labels.control-plane: "true"
  k8s:io.cilium.k8s.namespace.labels.kubernetes.azure.com/managedby: aks
  k8s:io.cilium.k8s.namespace.labels.kubernetes.io/cluster-service: "true"
  k8s:io.cilium.k8s.namespace.labels.kubernetes.io/metadata.name: kube-system
  k8s:io.cilium.k8s.policy.cluster: default
  k8s:io.cilium.k8s.policy.serviceaccount: coredns
  k8s:io.kubernetes.pod.namespace: kube-system
  k8s:k8s-app: kube-dns
  k8s:kubernetes.azure.com/managedby: aks
  k8s:version: v20
---
EOF
)
  done
  echo "$batch" | kubectl apply -f -
done

# Create CiliumEndpoint custom resources in batches
initial_num_cep="$(kubectl get ciliumendpoints --no-headers | wc -l || "0")"
for ((i=initial_num_cep; i<=$NUM_CILIUM_ENDPOINTS; i+=BATCH_SIZE)); do
  echo "Creating CiliumEndpoint batch $((i/BATCH_SIZE + 1))"
  batch=""
  for ((j=i; j<i+BATCH_SIZE && j<=$NUM_CILIUM_ENDPOINTS; j++)); do
    cep_name=$(printf "%06d" $j)
    cid_name=$(printf "%06d" $((j % NUM_CILIUM_IDENTITIES)))
    batch+=$(cat <<EOF

apiVersion: cilium.io/v2
kind: CiliumEndpoint
metadata:
  name: "$cep_name"
status:
  encryption: {}
  external-identifiers:
    container-id: 790d85075c394a8384f8b1a0fec62e2316c2556d175dab0c1fe676e5a6d92f33
    k8s-namespace: kube-system
    k8s-pod-name: coredns-54b69f46b8-dbcdl
    pod-name: kube-system/coredns-54b69f46b8-dbcdl
  id: 1453
  identity:
    id: $cid_name
    labels:
    - k8s:io.cilium.k8s.namespace.labels.addonmanager.kubernetes.io/mode=Reconcile
    - k8s:io.cilium.k8s.namespace.labels.control-plane=true
    - k8s:io.cilium.k8s.namespace.labels.kubernetes.azure.com/managedby=aks
    - k8s:io.cilium.k8s.namespace.labels.kubernetes.io/cluster-service=true
    - k8s:io.cilium.k8s.namespace.labels.kubernetes.io/metadata.name=kube-system
    - k8s:io.cilium.k8s.policy.cluster=default
    - k8s:io.cilium.k8s.policy.serviceaccount=coredns
    - k8s:io.kubernetes.pod.namespace=kube-system
    - k8s:k8s-app=kube-dns
    - k8s:kubernetes.azure.com/managedby=aks
    - k8s:version=v20
  named-ports:
  - name: dns
    port: 53
    protocol: UDP
  - name: dns-tcp
    port: 53
    protocol: TCP
  - name: metrics
    port: 9153
    protocol: TCP
  networking:
    addressing:
    - ipv4: 10.244.1.38
    node: 10.224.0.4
  policy:
    egress:
      enforcing: false
      state: <status disabled>
    ingress:
      enforcing: false
      state: <status disabled>
  state: ready
  visibility-policy-status: <status disabled>
---
EOF
)
  done
  echo "$batch" | kubectl apply -f -
done
