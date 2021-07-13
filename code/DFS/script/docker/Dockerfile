FROM golang:1.16-alpine
ENV GOPROXY=https://goproxy.cn \
    GO111MODULE=on
WORKDIR /DFS
COPY . /DFS/
RUN go build NodeRunner.go
