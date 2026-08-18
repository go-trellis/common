# common

Go utilities used across go-trellis services.

```bash
go get github.com/go-trellis/common
```

## Packages

| Package | Description |
|---------|-------------|
| [config](config/README.md) | JSON/YAML config, `#include` and `${}` substitution |
| [logger](logger) | logrus logger (`InitLogger`); `formatter`/`encoding`: json or text; file rotation |
| [txorm](txorm/README.md) | XORM engines, SQL logging, `Engine.Context` for `trace_id` |
| [transaction](transaction) | `Committer` (TX / NonTX) over named engines |
| [cache](cache/README.md) | LRU cache |
| [pool](pool) | Generic connection pool |
| [data-structures](data-structures) | Stack and queue |
| [clients/etcd](clients/etcd) | etcd client helpers |
| [crypto](crypto) | Hash, Base64, RC4, AES, TLS |
| [snowflake](snowflake/README.md) | Snowflake IDs |
| [errcode](errcode) | Structured error codes |
| [event](event) | In-process event bus |
| [plugin](plugin) / [injector](injector) | Plugins and dependency injection |
| [fsm](fsm/README.md) | Finite state machine |
| [types](types) | Type conversions and time helpers |
| [json](json) | JSON helpers |
| [files](files) | File I/O |
| [flagext](flagext) | Extra flag types |
| [shell](shell) | Command execution |
| [builder](builder/README.md) | Build-time version info |
| [assets](assets) | Embedded assets |
| [testutils](testutils) | Test assertions |

Logger `level` (logrus) and `databases.*.log_level` (xorm SQL) are independent. Bind request context with `engine.Context(ctx)` so SQL logs can include `trace_id`; do not `SetLogger` per request. Details: [txorm/README.md](txorm/README.md).

## Docs

- [config](config/README.md)
- [txorm](txorm/README.md)
- [cache](cache/README.md)
- [snowflake](snowflake/README.md)
- [fsm](fsm/README.md)
- [builder](builder/README.md)

## Also useful

- backoff: https://github.com/grafana/dskit/blob/main/backoff/backoff.go
- grpc limiter: https://github.com/grafana/dskit/tree/main/limiter
- go pool: https://github.com/Jeffail/tunny
