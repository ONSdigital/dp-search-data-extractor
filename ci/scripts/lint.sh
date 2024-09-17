#!/bin/bash -eux

pushd dp-search-data-extractor
  go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.60.1
  make lint
  npm install -g @asyncapi/cli
  make validate-specification
popd
