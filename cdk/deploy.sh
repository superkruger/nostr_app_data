#!/bin/bash

./build.sh
cdk deploy --context environment=$1 --profile PowerUserNostr --all
