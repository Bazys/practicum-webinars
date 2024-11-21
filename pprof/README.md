# example

## Профиль cpu

```sh
go tool pprof -http=":9090" -seconds=30 http://localhost:8080/debug/pprof/profile
```

## Память

```sh
go tool pprof -http=":9090" -seconds=30 http://localhost:8080/debug/pprof/heap
```

## codeprofile

```sh
curl localhost:8080/profile
go tool pprof -http=":9090" mem.profile
go tool pprof -http=":9090" cpu.profile
```

## mutex

[pprof](http://127.0.0.1:8081/debug/pprof/)

## labels

```sh
go tool pprof -seconds=30 http://localhost:8080/debug/pprof/profile

curl localhost:8080/profile/1
tags
tagfocus=handler:1
top
tagfocus=
tagignore=handler:1
top
web
```

## trace

```sh
curl "http://localhost:8080/debug/pprof/trace?seconds=15" > trace.out
go tool trace -http='localhost:9090' trace.out
```
