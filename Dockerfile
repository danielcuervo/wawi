FROM golang:1.10-alpine

ADD . /go/src/code

WORKDIR /go/src/code

RUN apk update \
    && apk upgrade \
    && apk add git \
    && go get -u github.com/golang/dep/cmd/dep \
    && dep init

CMD ["go", "run", "main.go"]