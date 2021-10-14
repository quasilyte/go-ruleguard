FROM golang:1.16-alpine AS build

ARG release
COPY . /root

RUN /root/test.bash $release
