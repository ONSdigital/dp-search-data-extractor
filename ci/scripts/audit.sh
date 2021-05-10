#!/bin/bash -eux

export cwd=$(pwd)

pushd $cwd/dp-search-data-extractor
  make audit
popd