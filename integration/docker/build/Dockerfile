FROM golang:1.11

VOLUME ["/go"]

WORKDIR /go/src/github.com/pandoprojects/pando/

ENV GOPATH=/go

ENV CGO_ENABLED=1 

ENV GO111MODULE=on

CMD ["/go/src/github.com/pandoprojects/pando/integration/docker/build/start.sh"]



