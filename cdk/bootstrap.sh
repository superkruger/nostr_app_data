#!/bin/bash

npx cdk bootstrap aws://418272791745/us-east-1 --profile AdminUserNostr --termination-protection --cloudformation-execution-policies arn:aws:iam::aws:policy/AdministratorAccess
