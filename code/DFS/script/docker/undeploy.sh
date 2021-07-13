docker rm -f $(docker ps -a |  grep "master*"  | awk '{print $1}')
docker rm -f $(docker ps -a |  grep "client*"  | awk '{print $1}')
docker rm -f $(docker ps -a |  grep "chunkServer*"  | awk '{print $1}')