# Kquetolk

A Redis compatible in-memory server implementing with RESP protocol.

## Testing

```sh
make buildapp # Build the application

make runapp # Run the application
```

```sh
redis-cli -p 6379 ping
# -> PONG

redis-cli -p 6379 echo hello
# -> "hello"
```
