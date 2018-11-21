#!/bin/bash

echo "==> Running go fmt"
go fmt ./... || (echo "go fmt updated formatting. Please commit formatted code first" ; exit 1)

echo "==> Running go vet"
go tool vet 2>/dev/null
if [ $? -eq 3 ]; then
    go get golang.org/x/tools/cmd/vet; \
fi

st=0
for pkg in $(go list ./... | grep -v /vendor/); do
    echo "===> $pkg"
    go vet "$pkg"
    [ $? -ne 0 ] && st=1
done

echo "==> Running go test"
go test -v ./... || (echo "go test failed. Please fix the test and try again" ; exit 1)

exit $st
