#!/usr/bin/env bash

packagePrefix=github.com/sandstorm/mailer-daemon

for subPackage in $(find . -d 1 -type d | grep -v documentation | grep -v dist | sed 's/.\///g')
do
    package=$packagePrefix/$subPackage
    echo go test $package
    go test $package
done
