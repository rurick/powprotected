# run server:
```shell
$ cd build/server
$ docker build . -t pow-server 
$ docker run --name pow-server -p 8888:8888 pow-server  
```

#run client:
```shell
$ cd build/client
$ go run ./../../cmd/client/main.go

```