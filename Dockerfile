FROM golang:1.18-bullseye AS builder

RUN mkdir -p /build/

COPY . /raft
WORKDIR /raft

RUN GOOS=linux GOARCH=amd64 go build -o /build/raft /raft/*.go

FROM ubuntu:jammy

RUN apt-get update && \
  apt-get install --no-install-recommends -y \
  ca-certificates \
  curl \
  netcat \
  tcpdump \
  net-tools \
  sudo

COPY --from=builder /build/ /usr/local/bin/
