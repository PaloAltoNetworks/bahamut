#!/bin/bash

set -x

echo "Cleaning test folder"
rm -rf test

mkdir -p test/results

echo "Installing dependencies"
make install_dependencies

echo "Lauching the tests"
make test &> test/results.txt

test_result=$?

cat test/results.txt | tee test/results/results.txt

sed 's/\[[0-9;]*m//g' test/results/results.txt > test/results/testresults.txt
go2xunit -fail -input test/results/testresults.txt -output test/results/testresults.xml
rm -rf test/results/results.txt
rm -rf test/results.txt

exit $test_result
