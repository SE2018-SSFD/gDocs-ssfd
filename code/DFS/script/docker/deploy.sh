r${a} dfsnode ./NodeRunner chunkServer 0.0.0.0:"${a}" ck"${a}" ${MASTER1PORT} &
done
sleep 2
docker run --expose=${MASTER1PORT} --name=master1 dfsnode ./NodeRunner master 0.0.0.0:${MASTER1PORT} ${MASTER1DIR} &
docker run --expose=${MASTER2PORT} --name=master2 dfsnode ./NodeRunner master 0.0.0.0:${MASTER2PORT} ${MASTER2DIR} &
docker run --expose=${MASTER3PORT} --name=master3 dfsnode ./NodeRunner master 0.0.0.0:${MASTER3PORT} ${MASTER3DIR} &

CLIENTPORT1=1333
CLIENTPORT2=1334
CLIENTPORT3=1335
docker run --expose=${CLIENTPORT1} --name=client1 dfsnode ./NodeRunner client 0.0.0.0:${CLIENTPORT1} ${MASTER1PORT} &
docker run --expose=${CLIENTPORT2} --name=client2 dfsnode ./NodeRunner client 0.0.0.0:${CLIENTPORT2} ${MASTER1PORT} &
docker run --expose=${CLIENTPORT3} --name=client3 dfsnode ./NodeRunner client 0.0.0.0:${CLIENTPORT3} ${MASTER1PORT} &