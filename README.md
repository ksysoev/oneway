# OneWay
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
