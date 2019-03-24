#!/bin/bash

OUT=benchmark.$(date +%y%m%d.%H%M%S)
cleanup() {
  rm -rf ${OUT}
}
cd $(dirname "${BASH_SOURCE[0]}")
trap cleanup EXIT
export GOPATH=
COUNT=2
go test --bench '.' ./ --benchmem --count ${COUNT} >> ${OUT}
go test --tags=btree --bench '.' ./ --benchmem --count ${COUNT} >> ${OUT}
go test --tags=llrb --bench '.' ./ --benchmem --count ${COUNT} >> ${OUT}

benchstat ${OUT}
