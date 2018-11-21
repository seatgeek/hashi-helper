#!/bin/bash

go tool vet 2>/dev/null
if [ $? -eq 3 ]; then
    go get golang.org/x/tools/cmd/vet; \
fi

st=0
for pkg in $(go list ./... | grep -v /vendor/); do
    echo "==> go vet $pkg"
    go vet "$pkg"
    [ $? -ne 0 ] && st=1
done

exit $st
