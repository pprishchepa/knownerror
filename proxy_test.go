package knownerror

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	t.Parallel()

	err := New("some validation error")
	require.NotNil(t, err)
	require.Equal(t, "some validation error", err.Error())
}

func TestNewf(t *testing.T) {
	t.Parallel()

	err := Newf("error: %s %d", "some code", 8234)
	require.NotNil(t, err)
	require.Equal(t, "error: some code 8234", err.Error())
}

func TestWrap(t *testing.T) {
	t.Parallel()

	base := errors.New("some base error")
	wrapped := Wrap(base)
	require.NotNil(t, wrapped)
	require.Equal(t, "some base error", wrapped.Error())
}

func TestWrap__nil(t *testing.T) {
	t.Parallel()

	wrapped := Wrap(nil)
	require.Nil(t, wrapped)
}

func TestProxy_WithCause(t *testing.T) {
	t.Parallel()

	outer := New("some outer error")
	cause := errors.New("some root cause")
	result := outer.WithCause(cause)

	require.Same(t, cause, result.Cause())
	require.Equal(t, "some outer error", result.Error())
}

func TestProxy_WithCause__nil(t *testing.T) {
	t.Parallel()

	outer := New("some outer error")
	result := outer.WithCause(nil)
	require.Same(t, outer, result)
}

func TestProxy_WithCause__preserves_identity(t *testing.T) {
	t.Parallel()

	outer := New("some outer error")
	cause := errors.New("some cause")
	result := outer.WithCause(cause)

	require.True(t, errors.Is(result, outer))
}

func TestProxy_Extends(t *testing.T) {
	t.Parallel()

	base := New("some base error")
	ext := errors.New("some extension")
	result := base.Extends(ext)

	require.True(t, errors.Is(result, ext))
}

func TestProxy_Extends__multiple(t *testing.T) {
	t.Parallel()

	base := New("some base error")
	ext1 := errors.New("some first extension")
	ext2 := errors.New("some second extension")
	result := base.Extends(ext1, ext2)

	require.True(t, errors.Is(result, ext1))
	require.True(t, errors.Is(result, ext2))
}

func TestProxy_Extends__ignores_nil(t *testing.T) {
	t.Parallel()

	base := New("some base error")
	ext := errors.New("some extension")
	result := base.Extends(nil, ext, nil)

	require.True(t, errors.Is(result, ext))
}

func TestProxy_Extends__all_nil(t *testing.T) {
	t.Parallel()

	base := New("some base error")
	result := base.Extends(nil, nil)
	require.Same(t, base, result)
}

func TestProxy_Error(t *testing.T) {
	t.Parallel()

	err := New("some message")
	require.Equal(t, "some message", err.Error())
}

func TestProxy_Error__nil_base(t *testing.T) {
	t.Parallel()

	proxy := &Proxy{}
	require.Empty(t, proxy.Error())
}

func TestProxy_Unwrap(t *testing.T) {
	t.Parallel()

	base := errors.New("some base error")
	wrapped := Wrap(base)
	require.Same(t, base, wrapped.Unwrap())
}

func TestProxy_Cause(t *testing.T) {
	t.Parallel()

	err := New("some error")
	require.Nil(t, err.Cause())
}

func TestProxy_Cause__set(t *testing.T) {
	t.Parallel()

	outer := New("some outer error")
	cause := errors.New("some cause")
	result := outer.WithCause(cause)
	require.Same(t, cause, result.Cause())
}

func TestProxy_Is(t *testing.T) {
	t.Parallel()

	err := New("some error")
	require.False(t, err.Is(nil))
}

func TestProxy_Is__matches_extended(t *testing.T) {
	t.Parallel()

	base := New("some base error")
	ext := errors.New("some extension")
	result := base.Extends(ext)

	require.True(t, result.Is(ext))
}

func TestProxy_Is__not_matches_non_extended(t *testing.T) {
	t.Parallel()

	base := New("some base error")
	other := errors.New("some other error")
	require.False(t, base.Is(other))
}

func TestProxy_Is__nested_extends(t *testing.T) {
	t.Parallel()

	err1 := errors.New("some first error")
	err2 := errors.New("some second error")
	base := New("some base error").Extends(err1).Extends(err2)

	require.True(t, errors.Is(base, err1))
	require.True(t, errors.Is(base, err2))
}

func TestProxy_As(t *testing.T) {
	t.Parallel()

	customErr := &customError{code: 8234}
	base := New("some base error").Extends(customErr)

	var target *customError
	require.True(t, errors.As(base, &target))
	require.Equal(t, 8234, target.code)
}

func TestProxy_As__non_matching_type(t *testing.T) {
	t.Parallel()

	base := New("some base error")
	var target *customError
	require.False(t, errors.As(base, &target))
}

func TestProxy_Format(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		got  struct {
			err    *Proxy
			format string
		}
		want struct {
			output string
		}
	}{
		{
			name: "format_s",
			got: struct {
				err    *Proxy
				format string
			}{
				err:    New("some error"),
				format: "%s",
			},
			want: struct {
				output string
			}{
				output: "some error",
			},
		},
		{
			name: "format_v",
			got: struct {
				err    *Proxy
				format string
			}{
				err:    New("some error"),
				format: "%v",
			},
			want: struct {
				output string
			}{
				output: "some error",
			},
		},
		{
			name: "format_plus_v_no_cause",
			got: struct {
				err    *Proxy
				format string
			}{
				err:    New("some error"),
				format: "%+v",
			},
			want: struct {
				output string
			}{
				output: "some error",
			},
		},
		{
			name: "format_q",
			got: struct {
				err    *Proxy
				format string
			}{
				err:    New("some error"),
				format: "%q",
			},
			want: struct {
				output string
			}{
				output: `"some error"`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := fmt.Sprintf(tt.got.format, tt.got.err)
			require.Equal(t, tt.want.output, result)
		})
	}
}

func TestProxy_Format__plus_v_with_cause(t *testing.T) {
	t.Parallel()

	cause := errors.New("some root cause")
	err := New("some error").WithCause(cause)
	result := fmt.Sprintf("%+v", err)
	require.Equal(t, "some error (cause: some root cause)", result)
}

func TestProxy_Format__plus_v_nested_cause(t *testing.T) {
	t.Parallel()

	innerCause := errors.New("some inner cause")
	outerCause := New("some outer cause").WithCause(innerCause)
	err := New("some main error").WithCause(outerCause)
	result := fmt.Sprintf("%+v", err)
	require.Equal(t, "some main error (cause: some outer cause)", result)
}

type customError struct {
	code int
}

func (e *customError) Error() string {
	return "custom error"
}
