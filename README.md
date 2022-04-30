# rpc-relay

## Overview

`rpc-relay` contains 4 subpackages:

- `ingress` is the HTTP server which contains the entrypoint for the end user.
The endpoint accepts the HTTP JSON-RPC calls via a POST request and either
retrieves the cached result or forwards the request to `egress` using NATS.
- `egress` contains the NATS subscriber which listens to calls from `ingress`,
does some initial checks on the incoming data, and forwards the request to
`jrpcserver`, then replies to the NATS request from `ingress` with the response
obtained from `jrpcserver`.
- `jrpcserver` is the actual HTTP JSON-RPC server which listens to RPC calls
on a given HTTP endpoint and returns the result. The server has multiple
RPC modules and methods for testing, and configuration allows to expose certain
RPC modules.
- `relayutil` contains utility functions, structs and methods, such as configuration
struct.

## Building

```shell
./scripts/buildall
```

This will build (`go build`) the server executables for the:
- ingress proxy (`bin/ingress`)
- egress proxy (`bin/egress`)
- HTTP JSON-RPC server (`bin/jrpcserver`)

Alternatively, run `scripts/buildsrv` with either `ingress`, `egress`
or `jrpcserver` as an argument.

## Configuration

All components accept the `-configPath` command line argument which defaults to
`config.dev.json`. Note that the `buildall` command automatically transforms
the executable and default config paths to absolute paths to avoid
errors during startup, so be wary of where you're launching the executables
from if you're not using the `buildall` script.

### Configuration variables

#### jrpcserver

- `host`. Defaults to `localhost`
- `port`. Defaults to `8001`
- `rpcEndpointUrl`. The HTTP API endpoint this server will listen on for incoming
JSON-RPC requests. Defaults to `/rpc`
- `enabledRpcModules`. The RPC methods are grouped into modules since the package
used for creating the RPC server (`go-ethereum`) does this similarly, and it builds method names based on both
structs and their methods. See the `jrpcserver.services` package for more details.
    - The keys in this config key are the Golang structs representing services you
    want to expose (e.g. `calculateSum` refers to `type CalculateSum struct`)
    - The values for each key are exposed methods in a module. **TODO**: by default all
    eliglble methods are exposed by `go-ethereum` and are available for calling;
    an additional check is required here. See `egress.handleRPCRequest` for details.

#### ingress

- `host`. Defaults to `localhost`
- `port`. Defaults to `8000`
- `refreshCachedRequestThreshold`. Threshold value in **seconds**. If less time than threshold had passed before 
a cached request was retrieved, a new request is not made. Defaults to `5.0`
- `expireCachedRequestThreshold`. Threshold value in **seconds**. All cached requests are expired if 
they're stored in the cache for longer than threshold value. Defaults to `30.0`
- `natsCallWaitTimeout`. Timeout in **seconds** for NATS/RPC call to egress proxy. Defaults to `5.0`
- `invalidateCacheLoopSleepPeriod`. Run cache invalidation each N **seconds**. Defaults to `5.0`

#### egress

- `host`. Defaults to `localhost`
- `port`. Defaults to `8000`.

Both values are currently unused since the egress proxy operates via NATS and
does not expose any HTTP endpoints.

#### nats

- `serverUrl`. NATS server url. Defaults to `nats://localhost:4222`
- `subjectName`. NATS subject name for RPC calls. Defaults to `jrpc.*.*`, where
the first wildcard is the RPC module name and the second is the RPC method name in the module
- `queueName`. NATS queue name for RPC calls. Defaults to `jrpcQueue

## Running

Make sure the NATS server is running and is available for communication.

Either 
```shell
./scripts/startall &
```

Alternatively, run `scripts/startsrv` with either `ingress`, `egress`
or `jrpcserver` as an argument.

Or just run the executables in the `bin` folder in the following order:
- `bin/jrpcserver`
- `bin/egress`
- `bin/ingress`

### Docker

Alternatively, run `docker-compose up -d` using the provided `docker-compose.yaml`.

This configuration uses default `config.dev.json` for demonstration.

Containers are running with `network_mode: "host"` for convenience.
Further tweaks to the build and deployment process are possible.

## Stopping

```shell
./scripts/stopall
```

Alternatively just `killall` the `ingress`, `egress`, and `jrpcserver` processes,
or use the `killsrv` script.

## Testing

Either 
```shell
./scripts/runtests
```

Or run `go test ./...` in the project's root folder.