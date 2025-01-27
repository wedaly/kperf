#!/usr/bin/env bash

set -xe

../bin/kperf rg run --runner-image=telescope.azurecr.io/oss/kperf:v0.1.6 --runnergroup="file://$(pwd)/runnergroup-spec.yaml"
../bin/kperf rg result