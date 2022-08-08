# udpwstunnel

This tool used to tunnel UDP data through websocket.

# How to run

In server.

First start UDP server with `nc` command.

```
nc -e /bin/cat -k -u -l 127.0.0.1 8888
```

Then start `wstunnel` in server.

```
./wstunnel -tl ws://domain:port -c udp://127.0.0.1:8888
```

In client.

```
./wstunnel -tc ws://domain:port/udp -l udp://127.0.0.1:7777
```

Finally use `nc` to do test in client.

```
nc -u 127.0.0.1 7777
```

# GO

This is my first go program. 
