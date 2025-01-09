#!/bin/bash
encname=${1:-"bftest"}
testcase=${2:-"testset/network_params.yaml"}
docker compose -f mysql.yml up -d && kurtosis run --enclave $encname github.com/tsinghua-cel/ethereum-package --args-file $testcase