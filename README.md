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
make tools    # install pinned buf, protoc-gen-go, and protoc-gen-go-grpc
make proto    # regenerate gen/pb with buf
make lint     # lint the protobuf schema
make test     # test release versioning; build, vet, and test the Go module
make verify   # run the complete lint, generation, and test gate
```

Schema changes must preserve protobuf wire compatibility. Run
`make breaking` against `main` before publishing a change.

## Release

Run the `release` GitHub Actions workflow from `main`. It validates the schema,
regenerates the committed Go package, runs the full test gate, then creates an
annotated semantic-version tag and a GitHub Release.

By default, versions follow natural days in `Asia/Shanghai`:

- the first release is `v0.1.0`;
- another release on the same day increments patch (`v0.1.1`, `v0.1.2`);
- the first release on a later day increments minor and resets patch
  (`v0.2.0`).

Select `major_release` to promote a `v0.x.y` release to `v1.0.0`. Go modules
require the module and import paths to include `/vN` for `v2` and later, so the
workflow rejects an automatic `v2.0.0`; prepare that module-path migration in a
dedicated change first.

Go consumers select a release in their module dependency:

```bash
go get github.com/housegate/rewriter-proto@v0.1.0
```
