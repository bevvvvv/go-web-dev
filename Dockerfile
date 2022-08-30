FROM golang:1.18-alpine

# copy project
# WORKDIR is $GOPATH
COPY ./ ./go-web-dev
WORKDIR $GOPATH/go-web-dev
# config.json is loaded in this directory
VOLUME $GOPATH/go-web-dev/config

# install dependencies
RUN go build -o build/ go-web-dev & \
    mkdir -p build/images/galleries & \
    cp -r ./config build/ & \
    cp -r ./views build/
    # & \
    #rm build/views/*.go & \
    #mv build go-web-dev & \
    #tar -czf "go-web-dev_${GOOS}_${GOARCH}.tar.gz" go-web-dev

RUN echo $GOOS
RUN ls
