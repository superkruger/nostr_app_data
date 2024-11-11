#!/bin/bash

functions=("./functions/connect" "./functions/default" "./functions/disconnect" "./functions/event" "./functions/forward" "./functions/request" "./functions/close")

cd ../app

for i in "${!functions[@]}"
do
    function=${functions[$i]}
    GOCACHE=/tmp go mod tidy && GOCACHE=/tmp GOARCH=arm64 GOOS=linux go build -tags lambda.norpc -o ${function}/build/bootstrap ${function}
done