# Dummy HTTP Server

This is a simple HTTP server written in Go that responds with a JSON message containing request details.

## Features

- Responds with request details in JSON format
- Configurable listen address and response message
- Graceful shutdown on SIGINT and SIGTERM signals

## Usage

```bash
> ./dummyhttp
```

### Command-line Flags
- `-l` : Address to listen on (default: :8080)
- `-m` : Message to return (default: hello, world)
- `-h` : Show help

**Example**
```bash
> ./dummyhttp -l :9090 -m "Hello, Go!"
```