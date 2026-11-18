FROM golang:1.26-alpine AS build

ARG release
COPY . /root

RUN /root/test.bash $release
