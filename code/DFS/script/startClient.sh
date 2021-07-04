#!/bin/bash
MASTER1ADDR=127.0.0.1:1234
CLIENTPORT=1333
./NodeRunner client 127.0.0.1:${CLIENTPORT} ${MASTER1ADDR} &> ../log/clientOutput${CLIENTPORT}.log &