#!/bin/bash -eux

pushd dp-search-data-extractor
  make lint
  npm install -g @asyncapi/cli
  make validate-specification
popd
