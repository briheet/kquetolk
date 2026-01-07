# Kquetolk

A Redis compatible in-memory server implementing with RESP protocol from socket scratch.

## Testing

```sh
make buildapp # Build the application

make runapp # Run the application
```

```sh
# To test Simple string
{ echo -e '+OK\r\n'; sleep 1; } | nc localhost 6379
```
