#!/bin/bash
docker run --name redis-server \
    -p 6379:6379 \
    -v /home/node1/data/redis:/data \
    -d redis:latest \
    --appendonly yes