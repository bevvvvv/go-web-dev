FROM golang:1.18-alpine

# copy project
# WORKDIR is $GOPATH
COPY ./ ./go-web-dev
WORKDIR $GOPATH/go-web-dev
# config.json is loaded in this directory
VOLUME $GOPATH/go-web-dev/config

# install dependencies
RUN go build

# app port
EXPOSE 3000
CMD go run *.go