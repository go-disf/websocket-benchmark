# websocket-benchmark
WebSocket-benchmark is a WebSocket benchmark tool, similar to the HTTP benchmark tool Apache bench (ab). It is simple and easy to use.



## Install

```sh
go install go-disf/websocket-benchmark
```

## Document

use `-h` to read the document:

```sh
$./websocket-benchmark -h
./websocket-benchmark [options] -url [http://]hostname[:port]/path
Options are:
  -b    Send data use websocket binary mode
  -c int
        Number of multiple requests to make at a time (default 1)
  -h    Display usage information (this message)
  -k    Use HTTP KeepAlive feature (default true)
  -n int
        Number of requests to perform
  -url string
        File containing data to POST. 
```

## Examples



```sh
$./websocket-benchmark -c 1000 -n 100000 -url ws://127.0.0.1:8080/ws
Benchmarking 127.0.0.1 (be patient)
Completed 100000 requests
Completed 200000 requests
Completed 300000 requests
Completed 400000 requests
Completed 500000 requests
Completed 600000 requests
Completed 700000 requests
Completed 800000 requests
Completed 900000 requests
Completed 1000000 requests

Server Scheme:          ws
Server Hostname:        127.0.0.1
Server Port:            8080

Document Path:          /ws

Concurrency Level:      1000
Time taken for tests:   31.0906628s
Complete requests:      1000000
Failed requests:        0
Total data sent:        71000000 bytes
Total data recv:        73000000 bytes
Requests per second:    32164.00 [#/sec] (mean)
Time per request:       30.57 [ms] (mean)
Transfer rate:          2347.9717 [Kbytes/sec] received
                        2283.6438 kb/s sent

Connection Times (ms)
                                        min  mean[+/-sd] median   max
Connect:         0               0       0       0       7
Sent:            0               0       0       0       32
Recv:            0               31      7       29      166
Total:           0               31      7       29      166

Percentage of the requests served within a certain time (ms)
  50%    29
  66%    30
  75%    32
  80%    33
  90%    37
  95%    41
  98%    47
  99%    53
 100%    166 (longest request)
```


## License

MIT licensed. See the LICENSE file for details.