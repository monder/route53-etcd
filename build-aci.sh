#!/bin/sh

export CGO_ENABLED=0
export GOOS=linux
export GOARCH=amd64
export VERSION=v0.3.1

go build -ldflags '-extldflags "-static"'
acbuild begin
acbuild set-name monder.cc/route53-etcd
acbuild copy route53-etcd /bin/route53-etcd
acbuild copy /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
acbuild set-exec /bin/route53-etcd
acbuild label add version $VERSION
acbuild label add arch $GOARCH
acbuild label add os $GOOS
acbuild environment add ETCDCTL_ENDPOINT http://172.16.28.1:2379
acbuild annotation add authors "Aleksejs Sinicins <monder@monder.cc>"
acbuild write route53-etcd-${VERSION}-${GOOS}-${GOARCH}.aci --overwrite
acbuild end
