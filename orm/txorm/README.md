# txorm

XORM wrapper with transaction helpers, config-driven engines, SQL logging, and request-scoped `trace_id`.

```bash
go get github.com/go-trellis/common/orm/txorm
```

Drivers: MySQL (default), SQLite3.

## Two different "level" knobs

Do not treat these as the same setting:

| Key | Package | Values | Controls |
|-----|---------|--------|----------|
| `logger.level` | logrus | `debug\|info\|warn\|error\|…` | Application logs (`Infof`, …) |
| `databases.*.log_level` | xorm filter | `debug\|info\|warn\|error\|off` (legacy ints 0–4 still accepted) | Engine PING / SQL / sync only |

`log_level` does **not** silence application `Infof` on the same `LogrusLogger`. Omitted `log_level` defaults to `debug`. SQL lines from `AfterSQL` need `debug` or `info`.

## Logger config

Pass the `logger` subtree to `logger.NewLogrusLoggerWithConfig`, then inject that logger into the engines:

```yaml
logger:
  log_path: /var/log/app.log
  rotate_mode: day          # day | hour
  max_age: 7d
  rotation_time: 24h
  level: info               # logrus sink (not databases.*.log_level)
  formatter: json           # json | text
  report_caller: true
  std_printers:
    - stdout
  writer_levels:
    - debug
    - info
    - warn
    - error

databases:
  test:
    dsn: "user:pass@tcp(127.0.0.1:3306)/db?parseTime=true"
    show_sql: true
    log_level: info         # xorm SQL filter
    is_default: true
    max_idle_conns: 10
    max_open_conns: 100
```

Database-only keys (no `logger` block) are listed in [example/mysql.yaml.sample](example/mysql.yaml.sample). Pass that file (or the `databases` subtree) to `NewEnginesWithConfig`.

```go
ll, err := logger.NewLogrusLoggerWithConfig(cfg.GetValuesConfig("logger"))
engines, err := txorm.NewEnginesWithConfig(cfg.GetValuesConfig("databases"), ll)
```

## Request `trace_id` on SQL logs

xorm's engine logger is a shared field. Replacing it with `SetLogger` per request is a data race and will mix `trace_id`s.

Bind the request context instead:

```go
// If tracing middleware already put trace_id on the request:
sessAny, err := engine.Context(req.Context()).NewSession()
sess := sessAny.(*xorm.Session)
defer sess.Close()

rows, err := sess.QueryString("SELECT 1")
```

Or set it yourself:

```go
ctx := tracing.WithTraceID(context.Background(), "my-trace-id")
sessAny, err := engine.Context(ctx).NewSession()
```

`AfterSQL` copies `trace_id` from that context (same key as `middleware/tracing.TraceIDKey`). `Ping` / application `Infof` have no session ctx and will not get `trace_id`.

Also valid:

```go
err := txorm.TransactionDo(engine.Context(ctx), func(s *xorm.Session) error {
    _, err := s.QueryString("SELECT 1")
    return err
})
```

### `Context()` wrapper rules

- Do **not** `SetLogger` per request.
- Do **not** `Close()` the engine returned by `Context()`. That `Close` is a no-op; close the original `*txorm.XEngine` (or the map from `NewEnginesWithConfig`).
- `engine.Context(ctx).TransactionDo(...)` keeps ctx (it is not the bare `*XEngine` method).

## Hooks

```go
_ = txorm.AddHook(engine, txorm.NewMetricsHook("test", prometheus.DefaultRegisterer))
```

`MetricsHook` records `trellis_txorm_query_total` and `trellis_txorm_query_duration_seconds`. `AddHook` accepts `*xorm.Engine`, `*txorm.XEngine`, or `transaction.Engine`.
