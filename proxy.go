// Package knownerror provides error proxies that can extend multiple errors
// while maintaining compatibility with errors.Is and errors.As.
package knownerror

import (
	"errors"
	"fmt"
)

// Proxy wraps an error, allows it to match multiple sentinel errors via Is/As,
// and can hold a root cause error.
type Proxy struct {
	base    error
	cause   error
	extends []error
}

// New creates a Proxy with a simple text message.
func New(text string) *Proxy {
	return &Proxy{base: errors.New(text)}
}

// Newf creates a Proxy with a formatted message.
func Newf(format string, args ...any) *Proxy {
	return &Proxy{base: fmt.Errorf(format, args...)}
}

// Wrap converts an existing error into a Proxy. Returns nil if err is nil.
func Wrap(err error) *Proxy {
	if err == nil {
		return nil
	}
	return &Proxy{base: err}
}

// WithCause attaches a root cause error and preserves the original error identity:
//
//	var ErrUserNotFound = knownerror.New("user not found")
//	err := ErrUserNotFound.WithCause(sql.ErrNoRows)
//	errors.Is(err, ErrUserNotFound) // true
//	err.Cause()                     // sql.ErrNoRows
func (e *Proxy) WithCause(cause error) *Proxy {
	if cause == nil {
		return e
	}
	cpy := *e
	cpy.cause = cause
	cpy.extends = make([]error, 0, len(e.extends)+1)
	cpy.extends = append(cpy.extends, e)
	cpy.extends = append(cpy.extends, e.extends...)
	return &cpy
}

// Extends adds error categories. The Proxy will match all extended errors via errors.Is:
//
//	var ErrNotFound = errors.New("not found")
//	var ErrUserNotFound = knownerror.New("user not found").Extends(ErrNotFound)
//	errors.Is(ErrUserNotFound, ErrNotFound) // true
func (e *Proxy) Extends(errs ...error) *Proxy {
	nonNilErrs := make([]error, 0, len(errs))
	for _, err := range errs {
		if err != nil {
			nonNilErrs = append(nonNilErrs, err)
		}
	}
	if len(nonNilErrs) == 0 {
		return e
	}
	cpy := *e
	cpy.extends = make([]error, 0, len(e.extends)+len(nonNilErrs))
	cpy.extends = append(cpy.extends, e.extends...)
	cpy.extends = append(cpy.extends, nonNilErrs...)
	return &cpy
}

// Error returns the error message.
func (e *Proxy) Error() string {
	if e.base != nil {
		return e.base.Error()
	}
	return ""
}

// Unwrap is a hook for errors.Unwrap. Returns the base error.
func (e *Proxy) Unwrap() error {
	return e.base
}

// Cause returns the root cause error attached via WithCause.
func (e *Proxy) Cause() error {
	return e.cause
}

// Is is a hook for errors.Is. Reports whether any extended error matches target.
func (e *Proxy) Is(target error) bool {
	if target == nil {
		return false
	}
	for _, ext := range e.extends {
		if errors.Is(ext, target) {
			return true
		}
	}
	return false
}

// As is a hook for errors.As. Finds the first extended error that matches target.
func (e *Proxy) As(target any) bool {
	for _, ext := range e.extends {
		if errors.As(ext, target) {
			return true
		}
	}
	return false
}

// Format implements fmt.Formatter. With %+v, prints the error and cause:
//
//	err := knownerror.New("db error").WithCause(errors.New("connection refused"))
//	fmt.Printf("%+v", err) // db error (cause: connection refused)
func (e *Proxy) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') && e.cause != nil {
			_, _ = fmt.Fprintf(s, "%s (cause: %s)", e.Error(), e.cause)
			return
		}
		fallthrough
	case 's':
		_, _ = fmt.Fprint(s, e.Error())
	case 'q':
		_, _ = fmt.Fprintf(s, "%q", e.Error())
	}
}
