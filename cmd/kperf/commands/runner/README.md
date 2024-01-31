## runner subcommand

This subcommand can be used to run benchmark.

Before we run benchmark, we need to define load profile.

```YAML
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
```

Let's say the local file `/tmp/example-loadprofile.yaml`.

We can run benchmark by the following command:

```bash
$ # cd kperf repo
$ # please build binary by make build
$
$ bin/kperf -v 3 runner run --config /tmp/example-loadprofile.yaml
I0131 09:50:45.471008 2312418 schedule.go:96] "Setting" clients=1000 connections=100 rate=100 total=10 http2=true content-type="json"
{
  "total": 10,
  "duration": "1.021348144s",
  "errorStats": {
    "unknownErrors": [],
    "responseCodes": {},
    "http2Errors": {}
  },
  "totalReceivedBytes": 18802170,
  "percentileLatencies": [
    [
      0,
      0.82955958
    ],
    [
      0.5,
      0.846259049
    ],
    [
      0.9,
      1.000932855
    ],
    [
      0.95,
      1.006544717
    ],
    [
      0.99,
      1.006544717
    ],
    [
      1,
      1.006544717
    ]
  ]
}
```

Please checkout `kperf runner run -h` to see more options.

If you want to run benchmark in kubernetes cluster, please use `kperf rg`.
