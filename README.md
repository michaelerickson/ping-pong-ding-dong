# Ping Pong, Ding Dong

This is a foolish little program meant to help learn Kubernetes. The basic
idea is that you deploy 4 services into a Kubernetes cluster:

- `ping`
- `pong`
- `ding`
- `dong`

Each service is actually just this program acting according to the environment
variable `PPDD_MODE` - which should be set to one of the values above.

Each service calls another service with a message. The message includes the
calling service's name and miscellaneous information to facilitate debugging
and learning.

```mermaid
sequenceDiagram
    participant Ping
    participant Pong
    participant Ding
    participant Dong
    
    Ping->>Pong: ping
    Pong->>Ping: pong
    Ping->>Ding: ping
    Ding->>Ping: ding
    Ding->>Dong: ding
    Dong->>Ping: dong
```

The basic idea is that `ping` periodically calls out to `pong` and `ding`.
The `dong` service is opaque to `ping` - it is activated by `ding`.

Again, I said this is kind of stupid. It's meant to be just complicated enough 
to try some things out like upgrading individual services. I intend to extend
it to play around with service mesh, gRPC, observability tools, etc.

# Environment Variables

The program accepts parameters via various environment variables.

- `PPDD_MODE`: required, sets the type of service
- `HTTP_PORT`: (_optional_ `8080`) but useful to set if testing in a
   non-Kubernetes environment
- `NAMESPACE`: (_optional_ `default`) Kubernetes namespace (set to `localhost`
   if testing locally)
- `PING_SVC`: (_optional_ `ping`)
- `PING_PORT`: (_optional_ `8080`)
- `PONG_SVC`: (_optional_ `pong`)
- `PONG_PORT`: (_optional_ `8080`)

Inside a cluster, the services will find each other via:

```text
 {PING_SVC}.{NAMESPACE}.svc.cluster.local
```

If `NAMESPACE == localhost` then we assume you are just running the binaries
locally in different processes so the services will find each other via:

```text
  localhost:{PING_PORT}
```

# REST Interface

The program exports the following endpoints.

- `GET /`: display the mode and various information about the service
- `POST /`: send a [JSON encoded message to the service](#message-format) 
- `GET /health`: JSON encoded health check
- `POST /shutdown`: cleanly shutdown the service

# Message Format

The endpoint `PUT /` accepts JSON encoded messages in the following format:

```json
{
  "Msg": "<string>"
}
```

Where `<string>` is one of; `ping`, `pong`, `ding`, or `dong`.

# Testing

## Manual Tests

### Single Mode
```shell
# Start the program in a mode of your choice
PPDD_MODE=dong HTTP_PORT=8989 go run .

# In another terminal, run commands
curl localhost:8989/health
# {"Status":"OK"}
curl localhost:8989/
# Hello from: 127.0.0.1:8989:
#   mode: dong
#   operating system: darwin
# ...
curl -v -X POST localhost:8989/ \
  -H 'Content-Type: application/json' \
  -d '{"Msg": "ding"}'

curl -X POST localhost:8989/shutdown
# Shutting down
```

### Ping and Pong

```shell
# Terminal 1
PPDD_MODE=ping \
  HTTP_PORT=8080 \
  NAMESPACE=localhost \
  PING_PORT=8080 \
  PONG_PORT=8081 \
  go run .

# Terminal 2
PPDD_MODE=pong \
  HTTP_PORT=8081 \
  NAMESPACE=localhost \
  PING_PORT=8080 \
  PONG_PORT=8081 \
  go run .
```

## Unit Tests
To run the tests locally and check coverage.

```shell
go test -v -cover -coverprofile=c.out .
go tool cover -html=c.out

```
