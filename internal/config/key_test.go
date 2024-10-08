package config_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/seanenck/lockbox/internal/config"
)

func TestDefaultKey(t *testing.T) {
	t.Setenv("LOCKBOX_KEYMODE", "")
	t.Setenv("LOCKBOX_KEY", "test")
	if _, err := config.NewKey(config.IgnoreKeyMode); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	t.Setenv("LOCKBOX_KEYMODE", "")
	t.Setenv("LOCKBOX_KEY", "")
	if _, err := config.NewKey(config.DefaultKeyMode); err == nil || err.Error() != "key MUST be set in this key mode" {
		t.Errorf("invalid error: %v", err)
	}
}

func TestNewKeyErrors(t *testing.T) {
	t.Setenv("LOCKBOX_KEYMODE", "invalid")
	if _, err := config.NewKey(config.IgnoreKeyMode); err == nil || err.Error() != "unknown key mode: invalid" {
		t.Errorf("invalid error: %v", err)
	}
	t.Setenv("LOCKBOX_KEYMODE", "none")
	t.Setenv("LOCKBOX_KEY", "  test")
	if _, err := config.NewKey(config.IgnoreKeyMode); err == nil || err.Error() != "key can NOT be set in this key mode" {
		t.Errorf("invalid error: %v", err)
	}
	t.Setenv("LOCKBOX_KEYMODE", "ask")
	t.Setenv("LOCKBOX_KEY", "test")
	if _, err := config.NewKey(config.IgnoreKeyMode); err == nil || err.Error() != "key can NOT be set in this key mode" {
		t.Errorf("invalid error: %v", err)
	}
	t.Setenv("LOCKBOX_KEYMODE", "command")
	t.Setenv("LOCKBOX_KEY", "   ")
	if _, err := config.NewKey(config.IgnoreKeyMode); err == nil || err.Error() != "key MUST be set in this key mode" {
		t.Errorf("invalid error: %v", err)
	}
	t.Setenv("LOCKBOX_KEYMODE", "plaintext")
	t.Setenv("LOCKBOX_KEY", "")
	if _, err := config.NewKey(config.IgnoreKeyMode); err == nil || err.Error() != "key MUST be set in this key mode" {
		t.Errorf("invalid error: %v", err)
	}
	t.Setenv("LOCKBOX_INTERACTIVE", "yes")
	t.Setenv("LOCKBOX_KEYMODE", "ask")
	t.Setenv("LOCKBOX_KEY", "")
	if _, err := config.NewKey(config.IgnoreKeyMode); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	t.Setenv("LOCKBOX_INTERACTIVE", "no")
	if _, err := config.NewKey(config.IgnoreKeyMode); err == nil || err.Error() != "ask key mode requested in non-interactive mode" {
		t.Errorf("invalid error: %v", err)
	}
}

func TestAskKey(t *testing.T) {
	t.Setenv("LOCKBOX_KEYMODE", "")
	t.Setenv("LOCKBOX_KEY", "test")
	k, _ := config.NewKey(config.IgnoreKeyMode)
	if k.Ask() {
		t.Error("invalid ask key")
	}
	t.Setenv("LOCKBOX_KEYMODE", "ask")
	t.Setenv("LOCKBOX_KEY", "")
	t.Setenv("LOCKBOX_INTERACTIVE", "no")
	k, _ = config.NewKey(config.IgnoreKeyMode)
	if k.Ask() {
		t.Error("invalid ask key")
	}
	t.Setenv("LOCKBOX_KEYMODE", "ask")
	t.Setenv("LOCKBOX_KEY", "")
	t.Setenv("LOCKBOX_INTERACTIVE", "yes")
	k, _ = config.NewKey(config.IgnoreKeyMode)
	if !k.Ask() {
		t.Error("invalid ask key")
	}
	fxn := func() (string, error) {
		return "", errors.New("TEST")
	}
	_, err := k.Read(fxn)
	if err == nil || err.Error() != "TEST" {
		t.Errorf("invalid error: %v", err)
	}
	fxn = func() (string, error) {
		return "", nil
	}
	_, err = k.Read(fxn)
	if err == nil || err.Error() != "key is empty" {
		t.Errorf("invalid error: %v", err)
	}
	fxn = func() (string, error) {
		return "abc", nil
	}
	val, err := k.Read(fxn)
	if err != nil || val != "abc" {
		t.Errorf("invalid error: %v", err)
	}
}

func TestIgnoreKey(t *testing.T) {
	t.Setenv("LOCKBOX_KEYMODE", "ignore")
	t.Setenv("LOCKBOX_KEY", "test")
	if _, err := config.NewKey(config.IgnoreKeyMode); err != nil {
		t.Errorf("invalid error: %v", err)
	}
	t.Setenv("LOCKBOX_KEYMODE", "ignore")
	t.Setenv("LOCKBOX_KEY", "")
	if _, err := config.NewKey(config.IgnoreKeyMode); err != nil {
		t.Errorf("invalid error: %v", err)
	}
}

func TestReadErrors(t *testing.T) {
	k := config.Key{}
	if _, err := k.Read(nil); err == nil || err.Error() != "invalid function given" {
		t.Errorf("invalid error: %v", err)
	}
	fxn := func() (string, error) {
		return "", nil
	}
	if _, err := k.Read(fxn); err == nil || err.Error() != "invalid key given" {
		t.Errorf("invalid error: %v", err)
	}
}

func TestPlainKey(t *testing.T) {
	t.Setenv("LOCKBOX_KEYMODE", "plaintext")
	t.Setenv("LOCKBOX_KEY", "  test ")
	k, err := config.NewKey(config.IgnoreKeyMode)
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	fxn := func() (string, error) {
		return "", nil
	}
	val, err := k.Read(fxn)
	if err != nil || val != "test" {
		t.Errorf("invalid error: %v", err)
	}
}

func TestReadIgnoreOrNoKey(t *testing.T) {
	t.Setenv("LOCKBOX_KEYMODE", "ignore")
	t.Setenv("LOCKBOX_KEY", "test")
	k, err := config.NewKey(config.IgnoreKeyMode)
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	fxn := func() (string, error) {
		return "", nil
	}
	val, err := k.Read(fxn)
	if err != nil || val != "" {
		t.Errorf("invalid error: %v", err)
	}
	t.Setenv("LOCKBOX_KEYMODE", "ignore")
	t.Setenv("LOCKBOX_KEY", "")
	k, err = config.NewKey(config.IgnoreKeyMode)
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	val, err = k.Read(fxn)
	if err != nil || val != "" {
		t.Errorf("invalid error: %v", err)
	}
	t.Setenv("LOCKBOX_KEYMODE", "none")
	k, err = config.NewKey(config.IgnoreKeyMode)
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	val, err = k.Read(fxn)
	if err != nil || val != "" {
		t.Errorf("invalid error: %v", err)
	}
}

func TestCommandKey(t *testing.T) {
	t.Setenv("LOCKBOX_KEYMODE", "command")
	t.Setenv("LOCKBOX_KEY", "thisisagarbagekey")
	k, err := config.NewKey(config.IgnoreKeyMode)
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	fxn := func() (string, error) {
		return "", nil
	}
	_, err = k.Read(fxn)
	if err == nil || !strings.HasPrefix(err.Error(), "key command failed:") {
		t.Errorf("invalid error: %v", err)
	}
}
