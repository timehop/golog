golog
=====

Severity-based key/value logging replacement for Go's standard logger.

Outputs a message, along with a set key/value pairs for easy parsing.

Example:

```go
// import "github.com/timehop/golog/log"

log.Error("Could not connect to server.", "url", url, "error", err.Error())
log.Info("Something happened.")
```

Would output something like:

```
ERROR | Could not connect to server. | url='http://timehop.com/', error='timed out'
INFO  | Something happened.
```
