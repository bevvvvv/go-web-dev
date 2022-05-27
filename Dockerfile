FROM golang:1.18-alpine

# copy project
# WORKDIR is $GOPATH
COPY ./ ./go-web-dev
WORKDIR $GOPATH/go-web-dev

# install dependencies
RUN go build

# app port
EXPOSE 3000
CMD go run main.go