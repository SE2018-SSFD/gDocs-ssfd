FROM golang:1.16-alpine
COPY . /backend/
ENV GOPROXY=https://goproxy.cn \
    GO111MODULE=on

WORKDIR /backend
CMD mkdir /mockdfs && go run main.go

EXPOSE 8080