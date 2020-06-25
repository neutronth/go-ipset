FROM golang:1.14 AS base-builder
WORKDIR /usr/src
COPY go.mod .
COPY go.sum .
RUN go mod download -x

FROM base-builder AS build-environment
COPY . .
