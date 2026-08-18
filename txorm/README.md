# txorm

XORM wrapper with transaction helpers, config-driven engines, SQL logging, and request-scoped `trace_id`.

```bash
go get github.com/go-trellis/common/txorm
```

Drivers: MySQL (default), SQLite3.

## Logger vs xorm log level

Do not treat these as the same setting:

| Key | Package | Values | Controls |
|-----|---------|--------|----------|
| `logger.level` | logrus | int 0–6 (`panic`…`trace`) | Application logs |
| `logger.formatter` / `logger.encoding` | logrus | `json` \| `text` (`formatter` wins if both set) | Log format |
| `databases.*.log_level` | xorm filter | int 0–4 (`debug`…`off`) | Engine PING / SQL / sync only |

`log_level` does **not** change the shared logrus sink. SQL from `AfterSQL` needs `0` or `1`.

## Logger config (`InitLogger`)

```yaml
logger:
  type: file                 # noop | console | file
  filename: /var/log/app.log
  level: 4                   # info
  formatter: json            # json | text (also: encoding)
  caller: true
  std_printers:
    - stdout
```

```go
ll, err := logger.InitLogger(cfg.GetValuesConfig("logger"))
engines, err := txorm.NewEnginesWithConfig(cfg.GetValuesConfig("databases"), logger.ToXormLogger(ll))
```

Database-only keys: [mysql.yaml.sample](mysql.yaml.sample).

## Request `trace_id` on SQL logs

xorm's engine logger is a shared field. Replacing it with `SetLogger` per request is a data race and will mix `trace_id`s.

Bind the request context instead. SQL `AfterSQL` reads `ctx.Value("trace_id")`:

```go
ctx := context.WithValue(req.Context(), "trace_id", traceID)

sessAny, err := engine.Context(ctx).NewSession()
sess := sessAny.(*xorm.Session)
defer sess.Close()

rows, err := sess.QueryString("SELECT 1")
```

`Ping` / application `Infof` have no session ctx and will not get `trace_id`.

```go
err := engine.Context(ctx).(interface {
    TransactionDo(func(*xorm.Session) error) error
}).TransactionDo(func(s *xorm.Session) error {
    _, err := s.QueryString("SELECT 1")
    return err
})
```

### `Context()` wrapper rules

- Do **not** `SetLogger` per request.
- Do **not** `Close()` the engine returned by `Context()`. That `Close` is a no-op; close the original `*txorm.XEngine`.
- `engine.Context(ctx).TransactionDo(...)` keeps ctx (it is not the bare `*XEngine` method).

## Hooks

```go
_ = txorm.AddHook(engine, txorm.NewMetricsHook("test", prometheus.DefaultRegisterer))
```

`MetricsHook` records `trellis_txorm_query_total` and `trellis_txorm_query_duration_seconds`.
