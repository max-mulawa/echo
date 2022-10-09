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
