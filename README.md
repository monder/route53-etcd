# route53-etcd
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

