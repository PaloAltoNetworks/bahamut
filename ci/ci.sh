#!/bin/bash

set -x

./ci/test.sh
if [ $? != 0 ]; then
    echo "test.sh failed"
    exit 1
fi

./ci/build.sh
if [ $? != 0 ]; then
    echo "build.sh failed"
    exit 1
fi
