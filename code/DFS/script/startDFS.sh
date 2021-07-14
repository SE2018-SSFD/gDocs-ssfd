#!/bin/bash
export GOPROXY=goproxy.cn

MASTER1ADDR=0.0.0.0:1234
MASTER2ADDR=0.0.0.0:1235
MASTER3ADDR=0.0.0.0:1236
MASTER1DIR=../log/log_1
MASTER2DIR=../log/log_2
MASTER3DIR=../log/log_3
CHUNKSERVERNUM=3
CHUNKSERVERROOT=../data/ck
rm -r ./NodeRunner
go build ../NodeRunner.go
rm -rf ${CHUNKSERVERROOT}
mkdir -p ${CHUNKSERVERROOT}

rm -rf ${MASTER1DIR}
mkdir ${MASTER1DIR}
rm -rf ${MASTER2DIR}
mkdir ${MASTER2DIR}
rm -rf ${MASTER3DIR}
mkdir ${MASTER3DIR}
for a in `eval echo {3000..$[CHUNKSERVERNUM+3000]}`
do
mkdir ${CHUNKSERVERROOT}/ck"${a}"
./NodeRunner chunkServer 0.0.0.0:"${a}" ${CHUNKSERVERROOT}/ck"${a}" ${MASTER1ADDR} &> ../log/chunkServerOutput"${a}".log &
done
sleep 2
./NodeRunner master ${MASTER1ADDR} ${MASTER1DIR}  &> ../log/masterOutput.log1 &
./NodeRunner master ${MASTER2ADDR} ${MASTER2DIR}  &> ../log/masterOutput.log2 &
./NodeRunner master ${MASTER3ADDR} ${MASTER3DIR}  &> ../log/masterOutput.log3 &

