#!/bin/bash
MASTER1ADDR=127.0.0.1:1234
MASTER2ADDR=127.0.0.1:1235
MASTER3ADDR=127.0.0.1:1236
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
./NodeRunner multimaster ${MASTER1ADDR} ${MASTER2ADDR} ${MASTER3ADDR} ${MASTER1DIR} ${MASTER2DIR} ${MASTER3DIR} &> ../log/masterOutput.log &
sleep 2
for a in `eval echo {3000..$[CHUNKSERVERNUM+3000]}`
do
mkdir ${CHUNKSERVERROOT}/ck"${a}"
./NodeRunner chunkServer 127.0.0.1:"${a}" ${CHUNKSERVERROOT}/ck"${a}" ${MASTER1ADDR} &> ../log/chunkServerOutput"${a}".log &
done

