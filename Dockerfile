FROM golang:1.5

ENV GO15VENDOREXPERIMENT=1
RUN go get github.com/mattn/gom

RUN mkdir -p /go/src/github.com/monder/route53-etcd
WORKDIR /go/src/github.com/monder/route53-etcd
COPY Gomfile /go/src/github.com/monder/route53-etcd/
RUN gom install

COPY . /go/src/github.com/monder/route53-etcd
RUN gom build app.go

ENTRYPOINT ["/go/src/github.com/monder/route53-etcd/app"]
