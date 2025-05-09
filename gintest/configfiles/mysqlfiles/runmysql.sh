#!/bin/bash

instances=$(docker ps -a | grep -i up | awk '{print $1}') # debug part
instances=$(docker ps -a | grep -i exited | awk '{print $1}')
# tr命令把输出中的换行符替换为空
instances=$(echo "$instances" | tr '\n' ' ')
# mapfile -t命令会把空格分隔的容器 ID 存储到数组docker_instances中
mapfile -t docker_instances <<< "$instances"
for id in "${docker_instances[@]}"
do  
    # docker stop $id # debug part
    docker rm $id
done
# 启动master节点
docker run --name mysql-master \
    -v ./master.my.cnf:/etc/mysql/my.cnf \
    -e MYSQL_ROOT_PASSWORD=123456 \
    -p 3306:3306 \
    -d mysql:8.0
# 启动slave节点
docker run --name mysql-slave \
    -v ./slave.my.cnf:/etc/mysql/my.cnf \
    -e MYSQL_ROOT_PASSWORD=123456 \
    -p 13306:3306 \
    -d mysql:8.0