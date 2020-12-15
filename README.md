# zlg (zerolog-gcp)

[![API reference](https://img.shields.io/badge/godoc-reference-5272B4)](https://pkg.go.dev/github.com/mark-ignacio/zerolog-gcp?tab=doc)
![GitHub](https://img.shields.io/github/license/mark-ignacio/zerolog-gcp)

**zlg** is a (hopefully) straightforward LevelWriter for using [zerolog](github.com/rs/zerolog) with Google Cloud Operations Logging, which used to be named Stackdriver.

Some notable features:

* The first log written to Cloud Logging is a slow, blocking write to confirm connectivity + permissions, but all subsequent writes are non-blocking.
* Handles converting `zerolog.WarnLevel` to `logging.Warning`.
* Zerolog's trace level maps to Cloud Logging's Default level.
* Cloud Logging's Alert and Emergency levels are not used.
* Ensure that all zlg-created loggers are flushed before program exit with `defer zlg.Flush()`

# Getting Started

## The usual cases

Logging only to Stackdriver:

```go
import zlg "github.com/mark-ignacio/zerolog-gcp"

// [...]

gcpWriter, err := zlg.NewCloudLoggingWriter(ctx, projectID, logID, zlg.CloudLoggingOptions{})
if err != nil {
    log.Panic().Err(err).Msg("could not create a CloudLoggingWriter")
}
log.Logger = log.Output(gcpWriter)
```

For non-GCP-hosted situations, you can log to both the console and GCP without much additional fuss.

```go
gcpWriter, err := zlg.NewCloudLoggingWriter(ctx, projectID, logID, zlg.CloudLoggingOptions{})
if err != nil {
    log.Panic().Err(err).Msg("could not create a CloudLoggingWriter")
}
log.Logger = log.Output(zerolog.MultiLevelWriter(
    zerolog.NewConsoleWriter(), 
    gcpWriter,
))
```

To ensure that the last asynchronous logs are delivered to Cloud Logging, zlg keeps a reference to all `logging.Logger` structs that zlg itself creates. If memory leaks of loggers are a concern, consider specifying providing a logger client via `zlg.CloudLoggingOptions` instead.

```go
gcpWriter, err := zlg.NewCloudLoggingWriter(ctx, projectID, logID, zlg.CloudLoggingOptions{})
if err != nil {
    log.Panic().Err(err).Msg("could not create a CloudLoggingWriter")
}
defer zlg.Flush()
doEverythingElse()
```

More advanced usage involves a non-empty [zlg.CloudLoggingOptions](https://pkg.go.dev/github.com/mark-ignacio/zerolog-gcp?tab=doc#CloudLoggingOptions).