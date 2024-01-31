## runnergroup subcommand

The subcommand is used to manage a set of runners in target kubernetes.
Before using command, please build container image for kperf first.

```bash
$ # cd kperf repo
$ # change repo name to your
$ export IMAGE_REPO=example.azurecr.io/public
$ export IMAGE_TAG=v0.0.2
$ make image-push
```

After that, build kperf binary.

```bash
$ # cd kperf repo
$ make build
```

> NOTE: `make help` can show more recipes.

### run - deploy a set of runners into kubernetes

Before run `run` command, we should define runner group first.
Here is an example: there are 10 runners in one group.

```YAML
# count defines how many runners in the group.
count: 10 # 10 runners in this group
# loadProfile defines what the load traffic looks like.
# All the runners in this group will use the same load profile.
loadProfile:
  version: 1
  description: example profile
  spec:
    # rate defines the maximum requests per second (zero is no limit).
    rate: 100
    # total defines the total number of requests.
    total: 10
    # conns defines total number of individual transports used for traffic.
    conns: 100
    # client defines total number of HTTP clients.
    client: 1000
    # contentType defines response's content type. (json or protobuf)
    contentType: json
    # disableHTTP2 means client will use HTTP/1.1 protocol if it's true.
    disableHTTP2: false
    # pick up requests randomly based on defined weight.
    requests:
      # staleList means this list request with zero resource version.
      - staleList:
          version: v1
          resource: pods
          limit: 1000
        shares: 1000 # Has 50% chance = 1000 / (1000 + 1000)
      # quorumList means this list request without kube-apiserver cache.
      - quorumList:
          version: v1
          resource: pods
          limit: 1000
        shares: 1000 # Has 50% chance = 1000 / (1000 + 1000)

# nodeAffinity defines how to deploy runners into dedicated nodes which have specific labels.
nodeAffinity:
  node.kubernetes.io/instance-type:
    - Standard_D8s_v3
```

Let's say the local file `/tmp/example-runnergroup-spec.yaml`.

We can run this runner group by the following command:

```bash
$ # cd kperf repo
$ # change repo name to your
$ export IMAGE_REPO=example.azurecr.io/public
$ export IMAGE_TAG=v0.0.2
$ export IMAGE_NAME=$IMAGE_REPO/kperf:$IMAGE_TAG
$
$ bin/kperf rg run \
  --runner-image=$IMAGE_NAME \
  --runnergroup="file:///tmp/example-runnergroup-spec.yaml"
```

We use URI scheme to load runner group's spec.
For example, `file://absolute-path`. We also support read spec from configmap, `configmap://name?namespace=ns&specName=dataNameInCM`.
Please checkout `kperf rg run -h` to see more options.

> NOTE: Currently, we use helm release to deploy a long running sever as controller to
deploy runners. The namespace is `runnergroups-kperf-io` and we don't allow run
multiple long running servers right now.

### status - check runnergroup's status

After deploy runner groups successfully, we can use `status` to check.

```bash
$ # cd kperf repo
$
$ bin/kperf rg status
NAME                   COUNT   SUCCEEDED   FAILED   STATE      START
runnergroup-server-0   10      10          0        finished   2024-01-30T10:18:36Z
```

### result - wait for test report

We use `result` to wait for report.

```bash
$ # cd kperf repo
$
$ bin/kperf rg result --wait --timeout=1h
{
  "total": 100,
  "duration": "318.47949ms",
  "errorStats": {
    "unknownErrors": [],
    "responseCodes": {},
    "http2Errors": {}
  },
  "totalReceivedBytes": 89149672,
  "percentileLatencies": [
    [
      0,
      0.039138658
    ],
    [
      0.5,
      0.072110663
    ],
    [
      0.9,
      0.158119337
    ],
    [
      0.95,
      0.179047998
    ],
    [
      0.99,
      0.236420101
    ],
    [
      1,
      0.267788626
    ]
  ]
}
```

`--wait` is used to block until all the runners finished.

### delete - delete runner groups

```bash
$ # cd kperf repo
$
$ bin/kperf rg delete
```
