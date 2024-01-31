# virtualcluster subcommand

## nodepool subcommand

The `nodepool` subcmd is using [kwok](https://github.com/kubernetes-sigs/kwok) to
deploy virtual nodepool. The user can use few physical resources to simulate
more than 1,000 nodes scenario.

The kperf uses `virtualnodes-kperf-io` namespace to host resources related to
nodepool.

If the user wants to schedule pods to virtual nodes, the user needs to change
node affinity and tolerations for pods.

```YAML
affinity:
  nodeAffinity:
    requiredDuringSchedulingIgnoredDuringExecution:
      nodeSelectorTerms:
      - matchExpressions:
        - key: type
          operator: In
          values:
          - kperf-virtualnodes

tolerations:
- key: "kperf.io/nodepool"
  operator: "Exists"
  effect: "NoSchedule"
```

Be default, the pod created by job controller will be completed after 5 seconds.
Other pods will be long running until receiving delete event.

### add - add a set of nodes with the same setting

We can use the following command to add nodepool named by `example` with 100 nodes.

```bash
$ # cd kperf repo
$ # please build binary by make build
$
$ bin/kperf vc nodepool add example \
  --nodes=100 --cpu=32 --memory=96 --max-pods=50 \
  --affinity="node.kubernetes.io/instance-type=Standard_D16s_v3"
```

The `--affinity` is used to deploy node controller (kwok) to nodes with the
specific labels.

The user can use `kubectl get nodes` to check.

```bash
$ kubectl get nodes -o wide | grep example | head -n 10
example-0                           Ready    agent   75s   fake      10.244.11.150   <none>        <unknown>            kwok-v0.4.0         kwok
example-1                           Ready    agent   75s   fake      10.244.9.71     <none>        <unknown>            kwok-v0.4.0         kwok
example-10                          Ready    agent   75s   fake      10.244.10.178   <none>        <unknown>            kwok-v0.4.0         kwok
example-11                          Ready    agent   75s   fake      10.244.9.74     <none>        <unknown>            kwok-v0.4.0         kwok
example-12                          Ready    agent   75s   fake      10.244.9.75     <none>        <unknown>            kwok-v0.4.0         kwok
example-13                          Ready    agent   75s   fake      10.244.11.143   <none>        <unknown>            kwok-v0.4.0         kwok
example-14                          Ready    agent   75s   fake      10.244.11.153   <none>        <unknown>            kwok-v0.4.0         kwok
example-15                          Ready    agent   75s   fake      10.244.10.180   <none>        <unknown>            kwok-v0.4.0         kwok
example-16                          Ready    agent   75s   fake      10.244.9.81     <none>        <unknown>            kwok-v0.4.0         kwok
example-17                          Ready    agent   75s   fake      10.244.11.147   <none>        <unknown>            kwok-v0.4.0         kwok
```

### list - list all the existing nodepools created by kperf

```bash
$ # cd kperf repo
$ # please build binary by make build
$
$ bin/kperf vc nodepool list
NAME         NODES     CPU   MEMORY (GiB)   MAX PODS   STATUS
example      ? / 100   32    96             50         deployed
example-v2   ? / 10    8     16             130        deployed
```

> NOTE: There is TODO item to show the number of ready nodes. Before that, we
use `?` as read nodes.

### delete - delete the target nodepool

```bash
$ # cd kperf repo
$ # please build binary by make build
$
$ bin/kperf vc nodepool delete example
$
$ bin/kperf vc nodepool list
NAME         NODES    CPU   MEMORY (GiB)   MAX PODS   STATUS
example-v2   ? / 10   8     16             130        deployed
```
