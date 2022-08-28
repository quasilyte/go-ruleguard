FROM golang:1.18-alpine AS build

ARG release
COPY . /root

RUN /root/test.bash $release
