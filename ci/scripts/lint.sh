#!/bin/bash -eux

pushd dp-search-data-extractor
  go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.5
  make lint
  npm install -g @asyncapi/cli
  npm install -g @redocly/cli
  make validate-specification
popd
