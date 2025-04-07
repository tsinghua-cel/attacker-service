#!/bin/bash
source env.sh
awk_output=$(kurtosis enclave inspect ${encname}| grep RUNNING | awk '{print $2}' | grep -v RUNNING)

services=()
while IFS= read -r line; do
    services+=("$line")
done <<< "$awk_output"
mkdir logs

for service in "${services[@]}"; do
        echo "dump $service logs"
        kurtosis service logs -a $encname $service > logs/${service}.log
done