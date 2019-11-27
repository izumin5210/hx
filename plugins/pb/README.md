# `pb` - Marshaling and Unmarshaling Protocol Buffers
[![GoDoc](https://godoc.org/github.com/izumin5210/hx/pb?status.svg)](https://godoc.org/github.com/izumin5210/hx/pb)

```go
err := hx.Post(ctx, "https://api.example.com/contents",
	pb.Proto(&in),
	hx.WhenSuccess(pb.AsProto(&out)),
	hx.WhenFailure(hx.AsError()),
)
```
