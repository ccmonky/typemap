# typemap

typemap contains a global TypeMap, which can be used to store various Types and their instances as resource registry.

## Requirements

- [Go 1.18 or newer](https://golang.org/dl/)

## Usage

```bash
go get -u github.com/ccmonky/typemap
```

```go
err := typemap.RegisterType[*Demo]()
err = typemap.Register[*Demo](ctx, "first", &Demo{})
demo, err = typemap.Get[*Demo](ctx, "first")
```
