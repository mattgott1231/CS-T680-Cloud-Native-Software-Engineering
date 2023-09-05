#!/bin/bash
docker build --tag votes-api-better:v1  -f ./dockerfile.better-votes .
docker build --tag voters-api-better:v1  -f ./dockerfile.better-voters .
docker build --tag polls-api-better:v1  -f ./dockerfile.better-polls .
