#!/bin/bash

./build.sh
cdk synth --context environment=$1
