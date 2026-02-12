# Go Structured Logging

**Version:** 0.1.0

**Tags:** logging, golang, observability, slog, structured-logging

**Scope:**
- **Languages:** go

## Table of Contents

- [Specification](#specification)
  - [Purpose](#purpose)
  - [Goals](#goals)
  - [Non-Goals](#non-goals)
- [Requirements](#requirements)
- [Golden Path](#golden-path)
  - [Steps](#steps)
  - [Examples](#examples)
    - [Logger bootstrap and configuration](#logger-bootstrap-and-configuration)
    - [Application logging and error handling](#application-logging-and-error-handling)
- [Good Examples](#good-examples)
  - [Typed attributes with proper message format](#typed-attributes-with-proper-message-format)
  - [Child logger for scoped context](#child-logger-for-scoped-context)
  - [Logger initialization in main](#logger-initialization-in-main)
  - [Error propagation with wrapping](#error-propagation-with-wrapping)
  - [Error logging at top level](#error-logging-at-top-level)
- [Bad Examples](#bad-examples)
  - [Bare key-value pairs without typed helpers](#bare-key-value-pairs-without-typed-helpers)
  - [Incorrect message formatting](#incorrect-message-formatting)
  - [Logging errors at intermediate layers](#logging-errors-at-intermediate-layers)
  - [Passing logger instance as function parameter](#passing-logger-instance-as-function-parameter)
- [References](#references)

## Specification

### Purpose

Define a consistent approach to structured logging in Go services using the standard library slog package, covering logger initialization via a factory pattern, message formatting conventions, and attribute usage.

### Goals

1. Ensure consistent structured logging across the project by using slog as the single logging library with typed attribute helpers.
2. Standardize logger initialization through a factory pattern with Config, CreateLoggerFromConfig, and slog.SetDefault in main.
3. Maintain readable log messages formatted as full sentences starting with a capital letter and ending with a period.
4. Enrich logs with contextual attributes using slog.String, slog.Int, and slog.Default().With for scoped child loggers.


### Non-Goals

1. Log aggregation infrastructure, log shipping, or external monitoring platform configuration.
2. Log file rotation, retention policies, or storage management.


## Requirements

### Rule 1: **MUST**

**Statement:** Use the standard library log/slog package as the sole logging library. Do not introduce third-party logging libraries such as logrus or zap.

**Rationale:** A single logging library ensures uniform log output across the entire codebase and eliminates the need to bridge between different logging APIs or output formats.

### Rule 2: **MUST**

**Statement:** Initialize the logger through the factory function CreateLoggerFromConfig and register it as the default logger via slog.SetDefault in the main function. Do not create loggers elsewhere.

**Rationale:** Centralizing logger creation in main guarantees a single configuration point for log level, format, and output destination, preventing scattered or conflicting logger setups across packages.

**Exceptions:**
- Creating a child logger with slog.Default().With() to add scoped contextual attributes within a specific function or request scope.

### Rule 3: **MUST**

**Statement:** Use package-level slog functions (slog.Info, slog.Debug, slog.Warn, slog.Error) for all log calls in application code outside of main. Do not store or pass *slog.Logger instances between functions.

**Rationale:** Package-level functions use the default logger set in main, keeping call sites simple and consistent. Passing logger instances adds unnecessary parameters and creates divergent logging paths.

**Exceptions:**
- Using a child logger returned by slog.Default().With() within the same function scope where it was created, to avoid repeating contextual attributes on every call.

### Rule 4: **MUST**

**Statement:** Format log messages as complete sentences that start with a capital letter and end with a period. Messages should describe a completed action or current state concisely.

**Rationale:** Consistent sentence-style messages improve log readability in both terminal output and log aggregation tools, making it easier to scan and filter log streams during debugging and operations.

### Rule 5: **MUST**

**Statement:** Use typed slog attribute helpers (slog.String, slog.Int, slog.Bool, slog.Any) for all log attributes. Do not use bare alternating key-value pairs like slog.Info("msg", "key", value).

**Rationale:** Typed helpers provide compile-time type safety, make attribute types explicit to the reader, and prevent subtle bugs from misaligned key-value pairs that the compiler cannot catch.

### Rule 6: **MUST**

**Statement:** Name all log attribute keys in camelCase. Do not use snake_case, kebab-case, or UPPER_CASE for attribute key names.

**Rationale:** Consistent camelCase key naming aligns with Go naming conventions, ensures uniform attribute keys across all log output, and simplifies querying and filtering in downstream log analysis systems.

### Rule 7: **MUST**

**Statement:** Propagate errors up the call stack using fmt.Errorf with the %w verb for wrapping. Do not log errors at intermediate layers. Log errors only at the top-level entry point in main.

**Rationale:** Logging errors at every layer produces duplicate and noisy log entries. Wrapping and propagating errors preserves the full context chain and lets the top-level handler decide on the appropriate log level and output.

**Exceptions:**
- The error is handled and recovered from at the current layer, and a warning or info log is appropriate to signal the recovery to operators without propagating the error further.

### Rule 8: **SHOULD**

**Statement:** Create child loggers using slog.Default().With() when a function makes multiple log calls that share the same contextual attributes, to avoid repeating those attributes on each call.

**Rationale:** Child loggers reduce attribute duplication across related log calls, keep code concise, and ensure consistent contextual enrichment for all log entries within a logical scope.

### Rule 9: **SHOULD**

**Statement:** Choose the appropriate log level based on audience and urgency. Use Debug for developer diagnostics, Info for operator-relevant state changes, Warn for recoverable anomalies, and Error only in main for fatal outcomes.

**Rationale:** Correct log level assignment enables effective filtering in production, keeps Info-level output actionable without noise, and reserves Error for situations that require immediate operator attention.



## Golden Path

### Steps

1. Configure the logger by populating the Config struct with the desired log level, output format, and quiet mode setting.
2. Create the logger in main by calling CreateLoggerFromConfig with the Config instance and register it as the process-wide default via slog.SetDefault.
3. Use package-level slog functions with typed attribute helpers for all log calls in application code outside of main.
4. Propagate errors up the call stack using fmt.Errorf with the %w verb and log errors only at the top-level entry point in main.


### Examples

#### Logger bootstrap and configuration

**When:**
- Starting a new Go application that requires structured logging with configurable level, format, and output destination.

**Steps:**
1. Define a Config struct instance choosing the log level from debug, info (default), warn, or error depending on the environment.
2. Select the output format from text-color (default, colored terminal output), text-no-color (plain structured text), or json (structured JSON for log aggregation systems).
3. Set quiet to true only when log output must be completely suppressed, such as in automated test environments or silent CLI modes.
4. Implement the CreateLoggerFromConfig factory function that parses the config, selects the appropriate slog handler based on format, and returns a configured slog.Logger instance.
5. Call CreateLoggerFromConfig in the main function, handle the returned error by writing to stderr and exiting, then register the logger as the default with slog.SetDefault.


**`config-struct-definition`**
```
// Config controls the logger behavior.
// Level:  "debug" | "info" (default) | "warn" | "error"
// Format: "text-color" (default) | "text-no-color" | "json"
// Quiet:  true suppresses all output (sends to io.Discard)
type Config struct {
    Level  string // Controls minimum severity of emitted logs.
    Format string // Selects the handler: colored text, plain text, or JSON.
    Quiet  bool   // When true, discards all log output.
}
```

**`logger-factory-implementation`**
```
func CreateLoggerFromConfig(config Config) (*slog.Logger, error) {
    level, err := parseLogLevel(config.Level)
    if err != nil {
        return nil, fmt.Errorf("parse log level: %w", err)
    }

    output := io.Writer(os.Stderr)
    if config.Quiet {
        output = io.Discard
    }

    handler, err := createLogHandler(config.Format, output, level)
    if err != nil {
        return nil, fmt.Errorf("create log handler: %w", err)
    }

    return slog.New(handler), nil
}

func createLogHandler(format string, output io.Writer, level slog.Level) (slog.Handler, error) {
    switch strings.ToLower(strings.TrimSpace(format)) {
    case "text-no-color":
        return slog.NewTextHandler(output, &slog.HandlerOptions{Level: level}), nil
    case "text-color":
        options := slogcolor.DefaultOptions
        options.Level = level
        return slogcolor.NewHandler(output, options), nil
    case "json":
        return slog.NewJSONHandler(output, &slog.HandlerOptions{Level: level}), nil
    default:
        return nil, fmt.Errorf("unknown log format: %s", format)
    }
}
```

**`main-bootstrap`**
```
func main() {
    cfg := parseFlags()

    logger, err := log.CreateLoggerFromConfig(cfg.Log)
    if err != nil {
        fmt.Fprintf(os.Stderr, "create logger: %v\n", err)
        os.Exit(1)
    }
    slog.SetDefault(logger)

    if err := run(cfg); err != nil {
        slog.Error("Application failed.", slog.String("error", err.Error()))
        os.Exit(1)
    }
}
```


#### Application logging and error handling

**When:**
- Writing business logic in service or repository layers that needs structured log output and correct error propagation.

**Steps:**
1. Call package-level slog functions with typed helpers such as slog.String and slog.Int, formatting the message as a sentence starting with a capital letter and ending with a period.
2. Create a child logger with slog.Default().With() when a function makes multiple log calls that share the same contextual attribute to avoid repeating it on each call.
3. Wrap and return errors with fmt.Errorf using the %w verb instead of logging them at intermediate layers, preserving the full error chain for the top-level caller.
4. Log errors exclusively in main with slog.Error and a typed slog.String attribute containing err.Error(), ensuring a single log entry per error occurrence.


**`service-function-with-child-logger`**
```
func (s *OrderService) ProcessBatch(batchID string, orders []Order) error {
    logger := slog.Default().With(slog.String("batchId", batchID))
    logger.Info("Processing batch.", slog.Int("orderCount", len(orders)))

    for _, order := range orders {
        if err := s.repo.Save(order); err != nil {
            return fmt.Errorf("save order %s: %w", order.ID, err)
        }
    }

    logger.Info("Batch processed successfully.")
    return nil
}
```

**`top-level-error-handling`**
```
if err := app.Run(ctx); err != nil {
    slog.Error("Application failed.", slog.String("error", err.Error()))
    os.Exit(1)
}
```




## Good Examples

### Typed attributes with proper message format

```go
slog.Info("Order processed successfully.",
    slog.String("orderId", order.ID),
    slog.Int("itemCount", len(order.Items)),
    slog.String("customerEmail", order.Email),
)
```

**Reason:** Uses typed attribute helpers slog.String and slog.Int for compile-time type safety. The message starts with a capital letter, describes a completed action, and ends with a period. Attribute keys use camelCase.

### Child logger for scoped context

```go
logger := slog.Default().With(slog.String("requestId", reqID))
logger.Debug("Parsing request body.")
logger.Info("Request handled successfully.")
```

**Reason:** Creates a child logger with slog.Default().With() to attach the requestId attribute to all subsequent log calls within the scope, avoiding repetition of the same attribute on each call.

### Logger initialization in main

```go
logger, err := log.CreateLoggerFromConfig(cfg.Log)
if err != nil {
    fmt.Fprintf(os.Stderr, "create logger: %v\n", err)
    os.Exit(1)
}
slog.SetDefault(logger)
```

**Reason:** Initializes the logger through the factory function and registers it as the default with slog.SetDefault in main, establishing a single configuration point for the entire application.

### Error propagation with wrapping

```go
conn, err := db.Connect(ctx, dsn)
if err != nil {
    return fmt.Errorf("connect to database: %w", err)
}
```

**Reason:** Wraps and propagates the error up the call stack using fmt.Errorf with the %w verb instead of logging it at this intermediate layer. The top-level caller in main decides how to log the error.

### Error logging at top level

```go
if err := app.Run(ctx); err != nil {
    slog.Error("Application failed.", slog.String("error", err.Error()))
    os.Exit(1)
}
```

**Reason:** Logs the error only at the top-level entry point in main using slog.Error with a typed slog.String attribute. This is the single place where errors are logged, keeping lower layers free of duplicate log entries.



## Bad Examples

### Bare key-value pairs without typed helpers

```go
slog.Error("Request failed.",
    "method", r.Method,
    "path", r.URL.Path,
    "status", statusCode,
)
```

**Reason:** Uses bare alternating key-value pairs instead of typed helpers like slog.String and slog.Int. This loses compile-time type safety and risks misaligned pairs that produce garbled log output at runtime without any compiler warning.

### Incorrect message formatting

```go
slog.Info("file uploaded", "name", fileName)
```

**Reason:** The message starts with a lowercase letter and lacks a trailing period, violating the sentence formatting convention. It also uses a bare key-value pair instead of slog.String for the attribute.

### Logging errors at intermediate layers

```go
conn, err := db.Connect(ctx, dsn)
if err != nil {
    slog.Error("Failed to connect.", slog.String("error", err.Error()))
    return fmt.Errorf("connect to database: %w", err)
}
```

**Reason:** Logs the error and then propagates it, causing the same error to appear in logs multiple times as each layer adds its own log entry. Errors should be wrapped and propagated, then logged only in main.

### Passing logger instance as function parameter

```go
func sendEmail(logger *slog.Logger, to string, body string) error {
    logger.Info("Sending email.", slog.String("recipient", to))
    return nil
}
```

**Reason:** Passes a logger instance as a function parameter instead of using package-level slog functions. This adds unnecessary parameters to function signatures and creates divergent logging paths across the codebase.



## References

| Title | Type | Link |
|-------|------|------|
| Go log/slog package documentation | documentation | https://pkg.go.dev/log/slog |
| slog-color library for colored text output | documentation | https://github.com/dusted-go/logging |

---

*This document was automatically generated by projectkit on 2026-02-12 23:12:03 CET. Do not edit manually.*
