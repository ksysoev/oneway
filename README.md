# OneWay

[![OneWay](https://github.com/ksysoev/oneway/actions/workflows/main.yml/badge.svg)](https://github.com/ksysoev/oneway/actions/workflows/main.yml)
[![CodeCov](https://codecov.io/gh/ksysoev/oneway/graph/badge.svg?token=3KGTO1UINI)](https://codecov.io/gh/ksysoev/oneway)
[![Go Report Card](https://goreportcard.com/badge/github.com/ksysoev/oneway)](https://goreportcard.com/report/github.com/ksysoev/oneway)
[![Go Reference](https://pkg.go.dev/badge/github.com/ksysoev/oneway.svg)](https://pkg.go.dev/github.com/ksysoev/oneway)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)

OneWay is network for connecting micro services via only outbound connectionss

Example of usage:

Start `exchange`, `revproxy` and example services(grpc + http)

```sh
docker compose up
```


make http request:

```sh 
go run example/http_client/main.go
```

make grpc request

```sh
go run example/grpc_client/main.go
```
