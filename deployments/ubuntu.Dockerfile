FROM ubuntu:20.04

ENV TZ=Europe/Moscow
ENV PATH=$PATH
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

RUN apt-get update -y \
    && apt-get install -y apt-transport-https tar wget curl nano gnupg git zlib1g-dev libxml2-dev libzip-dev \
    && apt install -y protobuf-compiler

#Install go1.16
RUN curl https://dl.google.com/go/go1.20.linux-arm64.tar.gz --output /tmp/go1.20.linux-arm64.tar.gz \
    && rm -rf /usr/local/go && tar -C /usr/local -xzf /tmp/go1.20.linux-arm64.tar.gz

RUN export GOROOT=/usr/local/go
RUN export GOPATH=$HOME/go
RUN export PATH=$PATH:$GOROOT/bin:$GOPATH/bin

ENV PATH=$PATH:/usr/local/go/bin:/root/go/bin

RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28 \
    && go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2 \
    && go install github.com/daixiang0/gci@latest \
    && go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.51.1 \
    && go install mvdan.cc/gofumpt@latest

RUN  go env

RUN mkdir /app

WORKDIR /app