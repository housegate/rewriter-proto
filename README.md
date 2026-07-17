# rewriter-proto

Wire contract for the HouseGate ClickHouse SQL rewriter. The protobuf schema
and generated Go package live here so the native Go implementation, the gRPC
service, and their callers can share one versioned contract.

Consumers import:

```go
import pb "github.com/housegate/rewriter-proto/gen/pb"
```

## Layout

| Path | What |
|---|---|
| `proto/rewriter.proto` | Authoritative protobuf and gRPC contract. |
| `gen/pb` | Generated Go messages and gRPC stubs (committed). |

## Build and regenerate

```bash
make tools    # install pinned protoc-gen-go and protoc-gen-go-grpc
make proto    # regenerate gen/pb with buf
make lint     # lint the protobuf schema
make test     # build, vet, and test the generated Go module
```

Schema changes must preserve protobuf wire compatibility. Run
`make breaking` against `main` before publishing a change.
