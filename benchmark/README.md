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
Resty-8      1.38ms ±81%
Sling-8      85.5µs ±45%
Gentleman-8  94.1µs ± 8%
Gorequest-8  1.51ms ±81%
Grequests-8  85.3µs ±31%
Hx-8         87.8µs ±16%
NetHTTP-8    76.9µs ± 8%

name         alloc/op
Resty-8      30.8kB ± 6%
Sling-8      7.83kB ± 1%
Gentleman-8  16.1kB ± 0%
Gorequest-8  23.2kB ± 1%
Grequests-8  7.17kB ± 1%
Hx-8         8.32kB ± 0%
NetHTTP-8    6.68kB ± 1%

name         allocs/op
Resty-8         184 ± 1%
Sling-8         100 ± 0%
Gentleman-8     245 ± 0%
Gorequest-8     199 ± 0%
Grequests-8    87.0 ± 0%
Hx-8            120 ± 0%
NetHTTP-8      83.0 ± 0%
```

## GET with Query

```
name         time/op
Resty-8      1.49ms ±82%
Sling-8      91.1µs ±27%
Gentleman-8   103µs ± 8%
Gorequest-8  1.61ms ±85%
Grequests-8  89.4µs ±37%
Hx-8         88.6µs ±16%
NetHTTP-8    78.4µs ± 6%

name         alloc/op
Resty-8      32.2kB ± 3%
Sling-8      8.25kB ± 0%
Gentleman-8  17.9kB ± 0%
Gorequest-8  20.4kB ± 0%
Grequests-8  8.88kB ± 0%
Hx-8         9.73kB ± 0%
NetHTTP-8    7.02kB ± 0%

name         allocs/op
Resty-8         202 ± 1%
Sling-8         119 ± 0%
Gentleman-8     271 ± 0%
Gorequest-8     162 ± 0%
Grequests-8     115 ± 0%
Hx-8            137 ± 0%
NetHTTP-8      96.0 ± 0%
```
