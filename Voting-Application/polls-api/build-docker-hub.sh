#!/bin/bash
docker buildx create --use 
docker buildx build --platform linux/amd64,linux/arm64  -f ./dockerfile.hub . -t mattgott1231/polls-api:v1 --push