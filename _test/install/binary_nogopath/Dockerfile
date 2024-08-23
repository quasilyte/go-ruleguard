FROM golang:1.22-alpine AS build

ARG release
COPY . /root

RUN /root/test.bash $release
