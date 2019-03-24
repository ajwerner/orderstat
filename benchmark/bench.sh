#!/bin/bash

OUT=benchmark.$(date +%y%m%d.%H%M%S)
cleanup() {
  rm -rf ${OUT}.1
  rm -rf ${OUT}.2
}
trap cleanup EXIT
export GOPATH=
go test --bench .* ./ --benchmem --count 5 > ${OUT}.1
go test --tags=btree --bench .* ./ --benchmem --count 5 > ${OUT}.2
benchstat ${OUT}.1 ${OUT}.2
