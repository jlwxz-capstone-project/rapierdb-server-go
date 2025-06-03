#!/bin/bash

# check whether cargo is installed
if ! command -v cargo &> /dev/null; then
    echo "cargo could not be found"
    exit 1
fi

cd ./pkg/loro/loro-c-ffi
cargo build --release