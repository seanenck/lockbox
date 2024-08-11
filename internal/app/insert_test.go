package app_test

import (
	"bytes"
	"errors"
	"io"
	"testing"

	"github.com/seanenck/lockbox/internal/app"
	"github.com/seanenck/lockbox/internal/backend"
)

type (
	mockInsert struct {
		command     *mockCommand
		noTOTP      func() (bool, error)
		input       func() ([]byte, error)
		pipe        func() bool
		token       func() string
		interactive bool
	}
)

func newMockInsert(t *testing.T) *mockInsert {
	m := &mockInsert{}
	m.command = newMockCommand(t)
	return m
}

func (m *mockInsert) TOTPToken() string {
	return m.token()
}

func (m *mockInsert) IsPipe() bool {
	return m.pipe()
}

func (m *mockInsert) Input(interactive bool) ([]byte, error) {
	m.interactive = interactive
	return m.input()
}

func (m *mockInsert) Args() []string {
	return m.command.Args()
}

func (m *mockInsert) Writer() io.Writer {
	return &m.command.buf
}

func (m *mockInsert) Confirm(p string) bool {
	return m.command.Confirm(p)
}

func (m *mockInsert) IsNoTOTP() (bool, error) {
	return m.noTOTP()
}

func (m *mockInsert) Transaction() *backend.Transaction {
	return m.command.Transaction()
}

func TestInsertDo(t *testing.T) {
	m := newMockInsert(t)
	m.pipe = func() bool {
		return false
	}
	m.command.args = []string{"test/test2"}
	m.command.confirm = false
	m.input = func() ([]byte, error) {
		return nil, errors.New("failure")
	}
	m.command.buf = bytes.Buffer{}
	if err := app.Insert(m, app.SingleLineInsert); err == nil || err.Error() != "invalid input: failure" {
		t.Errorf("invalid error: %v", err)
	}
	m.command.confirm = false
	m.pipe = func() bool {
		return true
	}
	if err := app.Insert(m, app.SingleLineInsert); err == nil || err.Error() != "invalid input: failure" {
		t.Errorf("invalid error: %v", err)
	}
	m.input = func() ([]byte, error) {
		return []byte("TEST"), nil
	}
	m.command.confirm = true
	m.command.args = []string{"a/b/c"}
	if err := app.Insert(m, app.SingleLineInsert); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if m.command.buf.String() != "" {
		t.Error("invalid insert")
	}
	m.pipe = func() bool {
		return false
	}
	m.command.buf = bytes.Buffer{}
	if err := app.Insert(m, app.SingleLineInsert); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if m.command.buf.String() == "" {
		t.Error("invalid insert")
	}
	m.command.buf = bytes.Buffer{}
	m.command.args = []string{"test/test2/test1"}
	if err := app.Insert(m, app.SingleLineInsert); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if m.command.buf.String() == "" {
		t.Error("invalid insert")
	}
	m.command.confirm = false
	m.command.buf = bytes.Buffer{}
	m.command.args = []string{"test/test2/test1"}
	if err := app.Insert(m, app.SingleLineInsert); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if m.command.buf.String() != "" {
		t.Error("invalid insert")
	}
	m.interactive = false
	m.command.confirm = true
	m.command.buf = bytes.Buffer{}
	m.command.args = []string{"test/test2/test1"}
	if err := app.Insert(m, app.SingleLineInsert); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if m.command.buf.String() == "" || !m.interactive {
		t.Error("invalid insert")
	}
	m.interactive = false
	m.command.buf = bytes.Buffer{}
	m.command.args = []string{"test/test2/test1"}
	if err := app.Insert(m, app.MultiLineInsert); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if m.command.buf.String() == "" || m.interactive {
		t.Error("invalid insert")
	}
}
