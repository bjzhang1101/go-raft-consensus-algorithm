FROM golang:1.18-bullseye AS builder

RUN mkdir -p /build/

COPY . /raft
WORKDIR /raft

RUN GOOS=linux GOARCH=amd64 go build -o /build/raft /raft/*.go

FROM ubuntu:jammy

COPY --from=builder /build/ /usr/local/bin/
