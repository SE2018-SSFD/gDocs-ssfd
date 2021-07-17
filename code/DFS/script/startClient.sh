#!/bin/bash
MASTER1ADDR=0.0.0.0:1234
CLIENTPORT=1333
./NodeRunner client 0.0.0.0:${CLIENTPORT} ${MASTER1ADDR} &> ../log/clientOutput${CLIENTPORT}.log &