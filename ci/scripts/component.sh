#!/bin/bash -eux

cwd=$(pwd)

pushd $cwd/dp-search-data-extractor
  make test-component
popd
