# Benchmark

- [Resty](https://github.com/go-resty/resty)
- [Sling](https://github.com/dghubble/sling)
- [gentleman](https://github.com/h2non/gentleman)
- [GoRequest](https://github.com/parnurzeal/gorequest)
- [GRequests](https://github.com/levigross/grequests)

```
go test -bench . -benchmem -count 30 -timeout 30m > bench.log
benchstat bench.log
```

## POST with JSON

```
name         time/op
Resty-8      1.37ms ±97%
Sling-8      94.2µs ±31%
Gentleman-8   104µs ± 9%
Gorequest-8  1.41ms ±81%
Grequests-8  84.6µs ±12%
Hx-8         81.8µs ± 9%
NetHTTP-8    76.2µs ± 8%

name         alloc/op
Resty-8      32.0kB ± 5%
Sling-8      7.92kB ± 0%
Gentleman-8  16.4kB ± 1%
Gorequest-8  23.4kB ± 2%
Grequests-8  7.29kB ± 1%
Hx-8         7.89kB ± 1%
NetHTTP-8    6.71kB ± 1%

name         allocs/op
Resty-8         185 ± 0%
Sling-8         100 ± 0%
Gentleman-8     245 ± 0%
Gorequest-8     199 ± 0%
Grequests-8    87.0 ± 0%
Hx-8            110 ± 0%
NetHTTP-8      83.0 ± 0%
```

## GET with Query

```
name         time/op
Resty-8      1.41ms ±81%
Sling-8      85.3µs ±18%
Gentleman-8   102µs ± 6%
Gorequest-8  1.21ms ±87%
Grequests-8  82.8µs ± 5%
Hx-8         84.0µs ± 8%
NetHTTP-8    78.0µs ± 7%

name         alloc/op
Resty-8      32.4kB ± 4%
Sling-8      8.25kB ± 0%
Gentleman-8  17.9kB ± 0%
Gorequest-8  20.4kB ± 1%
Grequests-8  8.88kB ± 0%
Hx-8         9.15kB ± 0%
NetHTTP-8    7.03kB ± 0%

name         allocs/op
Resty-8         202 ± 1%
Sling-8         119 ± 0%
Gentleman-8     271 ± 0%
Gorequest-8     162 ± 0%
Grequests-8     115 ± 0%
Hx-8            128 ± 0%
NetHTTP-8      96.0 ± 0%
```
