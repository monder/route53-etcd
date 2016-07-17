# route53-etcd

[![Go Report Card](https://goreportcard.com/badge/github.com/monder/route53-etcd)](https://goreportcard.com/report/github.com/monder/route53-etcd)
[![license](https://img.shields.io/github/license/monder/route53-etcd.svg?maxAge=2592000&style=flat-square)]()
[![GitHub tag](https://img.shields.io/github/tag/monder/route53-etcd.svg?style=flat-square)]()

Exposing IPs registred in etcd to route53

#Running

```
docker run monder/route53-etcd --etcd-endpoints=http://10.0.1.10:4001
```

Will read the configuration in etcd path `/hosts`

#Example

```
etcdctl set /hosts/AAAAAAAAAA/test.domain.lan "/units/test-app/*/ip"
```
Will read all keys in etcd matching pattern `/units/test-app/*/ip` and will create/update route53 recordsets for ZoneID `AAAAAAAAAA` and domain `test.domain.lan`

If multiple keys match the pattern - route53 will have multiple addresses for the same domain.

