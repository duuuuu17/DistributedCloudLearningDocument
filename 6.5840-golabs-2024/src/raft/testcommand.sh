#!/bin/bash
err=0
for((i = 1; i <= 20; i++)); do
    echo "Running command: go tset -run 3C count: $i"
    go test -run 3C
    if [ $? -ne 0 ]; then
        ((err++))
        echo "第 $err 执行失败"
    fi
done
echo "本次测试， 失败:$err 次"