# Protohackers

https://protohackers.com

## Troubleshooting

```bash
sudo tcpdump 'port 8888' -w /tmp/dump-external-prime.dmp
```

## Echo
Simple [Echo](https://www.rfc-editor.org/rfc/rfc862.txt) server written in go. 
Solution to [Problem 0](https://protohackers.com/problem/0)

### Run
```bash
make build
./bin/echo
```

```bash
nc localhost 7777
echo  #type and enter
echo  #returned
```
## Prime
Prime number checker.
Solution to [Problem 1](https://protohackers.com/problem/1)

```bash
make build
./bin/prime
```

```bash
nc localhost 8888
{"method":"isPrime","number":17}  #type and enter
{"method":"isPrime","prime":true}  #returned
```

### Means to an End
Mean price calculator.
Solution to [Problem 2](https://protohackers.com/problem/2)

```bash
make build
./bin/means
```

### chat server
Chat servers allowing multiple clients to communicate.
Solution to [Problem 3](https://protohackers.com/problem/3)

```bash
make build
./bin/chat
```

```bash
nc localhost 8888
max  #type and press enter
#begin chat

# another
nc localhost 8888
charlie95  #type and press enter
#begin chat
```

### Key-value store

Key-value store over UDP 
Solution to [Problem 4](https://protohackers.com/problem/4)

```bash
make build
./bin/kvstore
```

```bash
nc -u localhost 9999
key1=val1 #type and press Ctrl+D to avoid sending new line character
key1 #type and press Ctrl+D

=>key1=val1 #returned
```

# Proxy

Simple proxy server 
Solution to [Problem 5](https://protohackers.com/problem/5)

```bash
make build
./bin/proxy
```

see Chat server for interaction with proxy on port 8887