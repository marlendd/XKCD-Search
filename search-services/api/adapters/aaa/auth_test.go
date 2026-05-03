package aaa_test

import (
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"yadro.com/course/api/adapters/aaa"
	"yadro.com/course/api/core"
)

func newAAA(t *testing.T) aaa.AAA {
	t.Helper()
	t.Setenv("ADMIN_USER", "admin")
	t.Setenv("ADMIN_PASSWORD", "secret")

	a, err := aaa.New(time.Minute, slog.Default())
	assert.NoError(t, err)
	return a
}

func TestNew(t *testing.T) {
	t.Run("no env vars", func(t *testing.T) {
		_, err := aaa.New(time.Minute, slog.Default())
		assert.Error(t, err)
	})

	t.Run("only user set", func(t *testing.T) {
		t.Setenv("ADMIN_USER", "admin")
		_, err := aaa.New(time.Minute, slog.Default())
		assert.Error(t, err)
	})

	t.Run("success", func(t *testing.T) {
		a := newAAA(t)
		assert.NotNil(t, a)
	})
}

func TestLogin(t *testing.T) {
	a := newAAA(t)

	t.Run("successful login", func(t *testing.T) {
		token, err := a.Login("admin", "secret")
		assert.NoError(t, err)
		assert.NotEmpty(t, token)
	})

	t.Run("wrong password", func(t *testing.T) {
		_, err := a.Login("admin", "wrong")
		assert.ErrorIs(t, err, core.ErrNotAuthorized)
	})

	t.Run("unknown user", func(t *testing.T) {
		_, err := a.Login("unknown", "secret")
		assert.ErrorIs(t, err, core.ErrNotAuthorized)
	})
}

func TestVerify(t *testing.T) {
	a := newAAA(t)

	t.Run("valid token", func(t *testing.T) {
		token, err := a.Login("admin", "secret")
		assert.NoError(t, err)

		err = a.Verify(token)
		assert.NoError(t, err)
	})

	t.Run("invalid token", func(t *testing.T) {
		err := a.Verify("invalid-token")
		assert.ErrorIs(t, err, core.ErrNotAuthorized)
	})

	t.Run("expired token", func(t *testing.T) {
		t.Setenv("ADMIN_USER", "admin")
		t.Setenv("ADMIN_PASSWORD", "secret")

		expired, err := aaa.New(-time.Minute, slog.Default())
		assert.NoError(t, err)

		token, err := expired.Login("admin", "secret")
		assert.NoError(t, err)

		err = a.Verify(token)
		assert.ErrorIs(t, err, core.ErrNotAuthorized)
	})
}