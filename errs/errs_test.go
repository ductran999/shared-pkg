package errs_test

import (
	stderrors "errors"
	"testing"

	"github.com/ductran999/shared-pkg/errs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Constructors(t *testing.T) {
	tests := []struct {
		name string
		got  *errs.AppError
		kind errs.Kind
		code errs.Code
	}{
		{
			name: "internal",
			got:  errs.Internal("BOOM", "boom"),
			kind: errs.KindInternal,
			code: "BOOM",
		},
		{
			name: "not found",
			got:  errs.NotFound("USER_NOT_FOUND", "user not found"),
			kind: errs.KindNotFound,
			code: "USER_NOT_FOUND",
		},
		{
			name: "validation",
			got:  errs.Validation("INVALID_PARAM", "invalid"),
			kind: errs.KindValidation,
			code: "INVALID_PARAM",
		},
		{
			name: "conflict",
			got:  errs.Conflict("OBJECT_ALREADY_EXISTED", "already existed"),
			kind: errs.KindConflict,
			code: "OBJECT_ALREADY_EXISTED",
		},
		{
			name: "unauthorized",
			got:  errs.Unauthorized("UNAUTHORIZED", "unauthorized"),
			kind: errs.KindUnauthorized,
			code: "UNAUTHORIZED",
		},
		{
			name: "forbidden",
			got:  errs.Forbidden("FORBIDDEN", "forbidden"),
			kind: errs.KindForbidden,
			code: "FORBIDDEN",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.kind, tc.got.Kind)
			assert.Equal(t, tc.code, tc.got.Code)
			assert.NotEmpty(t, tc.got.Error())
			assert.Nil(t, tc.got.Cause)
		})
	}
}

func Test_ValidationDetails(t *testing.T) {
	details := []errs.FieldError{{Field: "email", Message: "invalid"}}
	got := errs.Validation("INVALID_PARAM", "validation failed", details...)

	assert.Equal(t, details, got.Details)
	assert.Nil(t, got.Cause)
}

func Test_Cause(t *testing.T) {
	root := stderrors.New("root cause")
	got := errs.NotFound("USER_NOT_FOUND", "user not found", root)

	require.NotNil(t, got.Cause)
	assert.ErrorIs(t, got, root)
	assert.Equal(t, root, got.Unwrap())
}

func Test_From(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		wantOK   bool
		wantCode errs.Code
	}{
		{
			name:     "direct",
			err:      errs.NotFound("USER_NOT_FOUND", "user not found"),
			wantOK:   true,
			wantCode: "USER_NOT_FOUND",
		},
		{
			name:     "wrapped",
			err:      stderrors.Join(stderrors.New("context"), errs.Conflict("CONFLICT", "conflict")),
			wantOK:   true,
			wantCode: "CONFLICT",
		},
		{
			name:   "plain error",
			err:    stderrors.New("plain"),
			wantOK: false,
		},
		{
			name:   "nil",
			err:    nil,
			wantOK: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := errs.From(tc.err)
			assert.Equal(t, tc.wantOK, ok)
			if tc.wantOK {
				assert.Equal(t, tc.wantCode, got.Code)
			} else {
				assert.Nil(t, got)
			}
		})
	}
}

func Test_HasCode(t *testing.T) {
	err := errs.Validation("INVALID_PARAM", "invalid")

	assert.True(t, errs.HasCode(err, "INVALID_PARAM"))
	assert.False(t, errs.HasCode(err, "OTHER"))
	assert.False(t, errs.HasCode(stderrors.New("plain"), "INVALID_PARAM"))
}

func Test_KindString(t *testing.T) {
	tests := []struct {
		kind errs.Kind
		want string
	}{
		{kind: errs.KindInternal, want: "internal"},
		{kind: errs.KindNotFound, want: "not_found"},
		{kind: errs.KindValidation, want: "validation"},
		{kind: errs.KindConflict, want: "conflict"},
		{kind: errs.KindUnauthorized, want: "unauthorized"},
		{kind: errs.KindForbidden, want: "forbidden"},
		{kind: errs.Kind(99), want: "unknown"},
	}

	for _, tc := range tests {
		t.Run(tc.want, func(t *testing.T) {
			assert.Equal(t, tc.want, tc.kind.String())
		})
	}
}
