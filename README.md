# Kperf

Kperf is a benchmark tool for Kubernetes API server.

It's like [wrk](https://github.com/wg/wrk), but it's designed to generate load and measure latency for Kubernetes API server.

## Quick Start

To quickly get started with Kperf, follow these steps:

1. Run the command `make` to build the necessary dependencies.

2. Once the build is complete, execute the following command to start the benchmark:

```bash
bin/kperf -v 3 runner run --config examples/node10_job1_pod100.yaml
```

3. The benchmark will generate load and measure the performance of the Kubernetes API server. You will see the results displayed in the terminal, including the total number of requests, duration, error statistics, received bytes, and percentile latencies.

Feel free to adjust the configuration file (`examples/node10_job1_pod100.yaml`) according to your requirements.

## Contributing

This project welcomes contributions and suggestions.  Most contributions require you to agree to a
Contributor License Agreement (CLA) declaring that you have the right to, and actually do, grant us
the rights to use your contribution. For details, visit https://cla.opensource.microsoft.com.

When you submit a pull request, a CLA bot will automatically determine whether you need to provide
a CLA and decorate the PR appropriately (e.g., status check, comment). Simply follow the instructions
provided by the bot. You will only need to do this once across all repos using our CLA.

This project has adopted the [Microsoft Open Source Code of Conduct](https://opensource.microsoft.com/codeofconduct/).
For more information see the [Code of Conduct FAQ](https://opensource.microsoft.com/codeofconduct/faq/) or
contact [opencode@microsoft.com](mailto:opencode@microsoft.com) with any additional questions or comments.

## Trademarks

This project may contain trademarks or logos for projects, products, or services. Authorized use of Microsoft
trademarks or logos is subject to and must follow
[Microsoft's Trademark & Brand Guidelines](https://www.microsoft.com/en-us/legal/intellectualproperty/trademarks/usage/general).
Use of Microsoft trademarks or logos in modified versions of this project must not cause confusion or imply Microsoft sponsorship.
Any use of third-party trademarks or logos are subject to those third-party's policies.
