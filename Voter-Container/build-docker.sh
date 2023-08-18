#!/bin/bash
docker buildx create --use 
docker buildx build --platform linux/amd64,linux/arm64  -f ./dockerfile.voters . -t mattgott1231/voters-api:v1 --push