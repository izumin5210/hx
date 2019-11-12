# Benchmark

```
go test -bench . -benchmem -count 100 -timeout 30m > bench.log
benchstat bench.log
```

```
name         time/op
Resty-8      1.60ms ±109%
Sling-8       92.3µs ±35%
Gentleman-8    101µs ±21%
Gorequest-8   1.74ms ±75%
Grequests-8    119µs ±26%
Hx-8           125µs ±32%
NetHTTP-8      100µs ±55%

name         alloc/op
Resty-8       31.1kB ± 5%
Sling-8       7.84kB ± 1%
Gentleman-8   14.7kB ± 1%
Gorequest-8   23.2kB ± 2%
Grequests-8   7.19kB ± 1%
Hx-8          8.32kB ± 1%
NetHTTP-8     6.69kB ± 1%

name         allocs/op
Resty-8          185 ± 1%
Sling-8          100 ± 0%
Gentleman-8      220 ± 0%
Gorequest-8      199 ± 0%
Grequests-8     87.0 ± 0%
Hx-8             120 ± 0%
NetHTTP-8       83.0 ± 0%
```

- [Resty](https://github.com/go-resty/resty)
- [Sling](https://github.com/dghubble/sling)
- [gentleman](https://github.com/h2non/gentleman)
- [GoRequest](https://github.com/parnurzeal/gorequest)
- [GRequests](https://github.com/levigross/grequests)
