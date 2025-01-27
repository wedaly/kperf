#!/usr/bin/env bash

set -xe

# KPERF_IMAGE="telescope.azurecr.io/oss/kperf:v0.1.6"
KPERF_IMAGE="widalytest.azurecr.io/kperf:widalytest003"

../bin/kperf rg run --runner-image=$KPERF_IMAGE --runnergroup="file://$(pwd)/runnergroup-spec.yaml"
../bin/kperf rg result
