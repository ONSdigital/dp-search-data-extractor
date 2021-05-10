#!/bin/bash -eux

pushd dp-search-data-extractor
  make build
  cp build/dp-search-data-extractor Dockerfile.concourse ../build
popd
