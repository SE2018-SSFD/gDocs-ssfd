docker volume rm mockdfs
docker volume create mockdfs
docker build -f ./go_1.16/Dockerfile ../.. -t gdoc_backend
docker-compose up -d