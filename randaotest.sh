#!/bin/bash

./build/bin/randao -validator-list ./cmd/randao/validator.json -case 1 -state ./cmd/randao/state.json
./build/bin/randao -validator-list ./cmd/randao/validator.json -case 2 -state ./cmd/randao/state.json
./build/bin/randao -validator-list ./cmd/randao/validator.json -case 3 -state ./cmd/randao/state.json
