MASTER1PORT=1234
MASTER2PORT=1235
MASTER3PORT=1236
MASTER1DIR=../log/log_1
MASTER2DIR=../log/log_2
MASTER3DIR=../log/log_3
CHUNKSERVERNUM=3
CHUNKSERVERROOT=../data/ck
rm -rf ${CHUNKSERVERROOT}
mkdir -p ${CHUNKSERVERROOT}
rm -rf ${MASTER1DIR}
mkdir ${MASTER1DIR}
rm -rf ${MASTER2DIR}
mkdir ${MASTER2DIR}
rm -rf ${MASTER3DIR}
mkdir ${MASTER3DIR}

docker build -f ./Dockerfile ../.. -t DFSNode
for a in `eval echo {3000..$[CHUNKSERVERNUM+3000]}`
do
mkdir ${CHUNKSERVERROOT}/ck"${a}"
docker run --expose=${a} --name=chunkServer${a} DFSNode ./NodeRunner chunkServer 0.0.0.0:"${a}" ${CHUNKSERVERROOT}/ck"${a}" ${MASTER1PORT} &> ../log/chunkServerOutput"${a}".log &
done
sleep 2
docker run --expose=${MASTER1PORT} --name=master1 DFSNode ./NodeRunner master ${MASTER1PORT} ${MASTER1DIR}  &> ../log/master1Output.log &
docker run --expose=${MASTER2PORT} --name=master2 DFSNode ./NodeRunner master ${MASTER2PORT} ${MASTER2DIR}  &> ../log/master2Output.log &
docker run --expose=${MASTER3PORT} --name=master3 DFSNode ./NodeRunner master ${MASTER3PORT} ${MASTER3DIR}  &> ../log/master3Output.log &

CLIENTPORT1=1333
CLIENTPORT2=1334
CLIENTPORT3=1335
docker run --expose=${CLIENTPORT1} --name=client1 DFSNode ./NodeRunner client 0.0.0.0:${CLIENTPORT1} ${MASTER1PORT} &> ../log/clientOutput${CLIENTPORT1}.log &
docker run --expose=${CLIENTPORT2} --name=client2 DFSNode ./NodeRunner client 0.0.0.0:${CLIENTPORT2} ${MASTER1PORT} &> ../log/clientOutput${CLIENTPORT2}.log &
docker run --expose=${CLIENTPORT3} --name=client3 DFSNode ./NodeRunner client 0.0.0.0:${CLIENTPORT3} ${MASTER1PORT} &> ../log/clientOutput${CLIENTPORT3}.log &