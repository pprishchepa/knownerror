# knownerror

[![CI](https://github.com/pprishchepa/knownerror/actions/workflows/ci.yml/badge.svg)](https://github.com/pprishchepa/knownerror/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/pprishchepa/knownerror)](https://goreportcard.com/report/github.com/pprishchepa/knownerror)
[![codecov](https://codecov.io/gh/pprishchepa/knownerror/branch/main/graph/badge.svg)](https://codecov.io/gh/pprishchepa/knownerror)
[![Go Reference](https://pkg.go.dev/badge/github.com/pprishchepa/knownerror.svg)](https://pkg.go.dev/github.com/pprishchepa/knownerror)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A Go library for creating "known" errors that can extend other errors while maintaining compatibility with `errors.Is()` and `errors.As()`.

## When to use

**This library is not needed for most Go projects.** Standard `errors.New()`, `fmt.Errorf()`, and error wrapping cover the majority of use cases.

Consider this library when:
- You have many sentinel errors across different packages
- You want to categorize errors for mapping in the presentation layer (e.g., HTTP status codes, gRPC codes) without checking each specific error
- You need an error to match multiple categories via `errors.Is()`
- You want to add categories without creating custom error types
- You want to attach a root cause to sentinel errors while preserving `errors.Is()` matching

Example: instead of mapping each error individually:

```go
switch {
case errors.Is(err, user.ErrNotFound):
    return http.StatusNotFound
case errors.Is(err, order.ErrNotFound):
    return http.StatusNotFound
case errors.Is(err, product.ErrNotFound):
    return http.StatusNotFound
// ... many more
}
```

You can define error categories and check once:

```go
var ErrNotFound = errors.New("not found")

// In each package:
var ErrUserNotFound = knownerror.New("user not found").Extends(ErrNotFound)
var ErrOrderNotFound = knownerror.New("order not found").Extends(ErrNotFound)

// In presentation layer:
if errors.Is(err, ErrNotFound) {
    return http.StatusNotFound
}
```

Errors can belong to multiple categories:

```go
var (
    ErrNotFound  = errors.New("not found")
    ErrForbidden = errors.New("forbidden")
)

// User tried to access another user's deleted resource
var ErrNotFoundOrForbidden = knownerror.New("resource not available").Extends(ErrNotFound, ErrForbidden)

errors.Is(ErrNotFoundOrForbidden, ErrNotFound)  // true
errors.Is(ErrNotFoundOrForbidden, ErrForbidden) // true
```

## Installation

```bash
go get github.com/pprishchepa/knownerror
```

## Usage

### Creating a known error

```go
var ErrNotFound = knownerror.New("not found")
var ErrValidation = knownerror.Newf("validation failed: %s", "invalid input")
```

### Wrapping an existing error

```go
baseErr := errors.New("database connection failed")
wrapped := knownerror.Wrap(baseErr)
```

### Adding a cause error

Use `WithCause` to attach a root cause error while preserving the original error identity:

```go
var ErrUserNotFound = knownerror.New("user not found")

func GetUser(id string) (*User, error) {
    user, err := db.FindUser(id)
    if err != nil {
        return nil, ErrUserNotFound.WithCause(err)
    }
    return user, nil
}

// Later, check the error type:
if errors.Is(err, ErrUserNotFound) {
    // Handle user not found
}

// Access the root cause:
type Causer interface { Cause() error }
var causer Causer
if errors.As(err, &causer) {
    rootCause := causer.Cause()
}
```

### Extending with other errors

Use `Extends` to make an error match multiple sentinel errors:

```go
var (
    ErrNotFound   = knownerror.New("not found")
    ErrBadRequest = knownerror.New("bad request")
)

var ErrUserNotFound = knownerror.New("user not found").Extends(ErrNotFound, ErrBadRequest)

// Now both checks return true:
errors.Is(ErrUserNotFound, ErrNotFound)   // true
errors.Is(ErrUserNotFound, ErrBadRequest) // true
```

### Formatting with %+v

When using `%+v`, the error prints both the message and the cause:

```go
cause := errors.New("connection refused")
err := knownerror.New("database error").WithCause(cause)

fmt.Printf("%v\n", err)  // database error
fmt.Printf("%+v\n", err) // database error (cause: connection refused)
```

## API

### Functions

- `New(text string) *Proxy` - creates a new error with the given message
- `Newf(format string, args ...any) *Proxy` - creates a new formatted error
- `Wrap(err error) *Proxy` - wraps an existing error (returns nil if err is nil)

### Methods

- `WithCause(cause error) *Proxy` - returns a copy with a root cause error attached
- `Extends(errs ...error) *Proxy` - returns a copy that matches additional errors via `Is`/`As`
- `Error() string` - returns the error message
- `Unwrap() error` - returns the base error
- `Cause() error` - returns the root cause error (set via `WithCause`)
- `Is(target error) bool` - checks if any extended error matches the target
- `As(target any) bool` - extracts a matching extended error into the target
- `Format(s fmt.State, verb rune)` - implements `fmt.Formatter` for custom formatting

## License

MIT License - see [LICENSE](LICENSE) for details.
