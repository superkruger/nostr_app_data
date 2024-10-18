#!/bin/bash

ssh -i "~/.aws/nostr_ec2_keys.pem" -L 27017:NostrElasticDocDB-418272791745.us-east-1.docdb-elastic.amazonaws.com:27017 ec2-user@ec2-54-172-184-55.compute-1.amazonaws.com -N