# For testing
FROM golang:1.13
MAINTAINER Chris Sandison <chris@thinkdataworks.com>

ENV GO111MODULE on
ENV CGO_ENABLED 0
ENV GOOS linux
ENV PROJECT_PATH /app

WORKDIR $PROJECT_PATH

RUN apt-get update && \
    apt-get install -yqq zip

# build/cache modules
COPY go.mod $PROJECT_PATH
COPY go.sum $PROJECT_PATH
RUN go mod download

# install development dependencies
RUN go get github.com/rakyll/gotest && \
    go get github.com/stretchr/testify

# move project files
COPY . $PROJECT_PATH

# vendor modules (must be after the project files are copied)
RUN go mod vendor

