#!/bin/bash -eux

pushd dp-search-data-extractor
  go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.2.2
  make lint
  npm install -g @asyncapi/cli
  npm install -g @redocly/cli
  make validate-specification
popd
