package backend_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/seanenck/lockbox/internal/backend"
	"github.com/seanenck/lockbox/internal/platform"
)

const (
	testDir = "testdata"
)

func testFile(name string) string {
	file := filepath.Join(testDir, name)
	if !platform.PathExists(testDir) {
		os.Mkdir(testDir, 0o755)
	}
	return file
}

func fullSetup(t *testing.T, keep bool) *backend.Transaction {
	file := testFile("test.kdbx")
	if !keep {
		os.Remove(file)
	}
	os.Setenv("LOCKBOX_READONLY", "no")
	os.Setenv("LOCKBOX_STORE", file)
	os.Setenv("LOCKBOX_KEY", "test")
	os.Setenv("LOCKBOX_KEYFILE", "")
	os.Setenv("LOCKBOX_KEYMODE", "plaintext")
	os.Setenv("LOCKBOX_TOTP", "totp")
	os.Setenv("LOCKBOX_HOOKDIR", "")
	os.Setenv("LOCKBOX_SET_MODTIME", "")
	tr, err := backend.NewTransaction()
	if err != nil {
		t.Errorf("failed: %v", err)
	}
	return tr
}

func TestKeyFile(t *testing.T) {
	os.Clearenv()
	file := testFile("keyfile_test.kdbx")
	keyFile := testFile("file.key")
	os.Remove(file)
	os.Setenv("LOCKBOX_READONLY", "no")
	os.Setenv("LOCKBOX_STORE", file)
	os.Setenv("LOCKBOX_KEY", "test")
	os.Setenv("LOCKBOX_KEYFILE", keyFile)
	os.Setenv("LOCKBOX_KEYMODE", "plaintext")
	os.Setenv("LOCKBOX_TOTP", "totp")
	os.Setenv("LOCKBOX_HOOKDIR", "")
	os.Setenv("LOCKBOX_SET_MODTIME", "")
	os.WriteFile(keyFile, []byte("test"), 0o644)
	tr, err := backend.NewTransaction()
	if err != nil {
		t.Errorf("failed: %v", err)
	}
	if err := tr.Insert(backend.NewPath("a", "b"), "t"); err != nil {
		t.Errorf("no error: %v", err)
	}
}

func setup(t *testing.T) *backend.Transaction {
	return fullSetup(t, false)
}

func TestNoWriteOnRO(t *testing.T) {
	setup(t)
	os.Setenv("LOCKBOX_READONLY", "yes")
	tr, _ := backend.NewTransaction()
	if err := tr.Insert("a/a/a", "a"); err.Error() != "unable to alter database in readonly mode" {
		t.Errorf("wrong error: %v", err)
	}
}

func TestBadTOTP(t *testing.T) {
	tr := setup(t)
	os.Setenv("LOCKBOX_TOTP", "Title")
	if err := tr.Insert("a/a/a", "a"); err.Error() != "invalid totp field, uses restricted name" {
		t.Errorf("wrong error: %v", err)
	}
}

func TestBadAction(t *testing.T) {
	tr := &backend.Transaction{}
	if err := tr.Insert("a/a/a", "a"); err.Error() != "invalid transaction" {
		t.Errorf("wrong error: %v", err)
	}
}

func TestMove(t *testing.T) {
	setup(t)
	fullSetup(t, true).Insert(backend.NewPath("test", "test2", "test1"), "pass")
	fullSetup(t, true).Insert(backend.NewPath("test", "test2", "test3"), "pass")
	if err := fullSetup(t, true).Move(backend.QueryEntity{Path: backend.NewPath("test", "test2", "test3"), Value: "abc"}, backend.NewPath("test1", "test2", "test3")); err != nil {
		t.Errorf("no error: %v", err)
	}
	q, err := fullSetup(t, true).Get(backend.NewPath("test1", "test2", "test3"), backend.SecretValue)
	if err != nil {
		t.Errorf("no error: %v", err)
	}
	if q.Value != "abc" {
		t.Errorf("invalid retrieval")
	}
	if err := fullSetup(t, true).Move(backend.QueryEntity{Path: backend.NewPath("test", "test2", "test1"), Value: "test"}, backend.NewPath("test1", "test2", "test3")); err != nil {
		t.Errorf("no error: %v", err)
	}
	q, err = fullSetup(t, true).Get(backend.NewPath("test1", "test2", "test3"), backend.SecretValue)
	if err != nil {
		t.Errorf("no error: %v", err)
	}
	if q.Value != "test" {
		t.Errorf("invalid retrieval")
	}
}

func TestInserts(t *testing.T) {
	if err := setup(t).Insert("", ""); err.Error() != "empty path not allowed" {
		t.Errorf("wrong error: %v", err)
	}
	if err := setup(t).Insert("tests", "test"); err.Error() != "input paths must contain at LEAST 2 components" {
		t.Errorf("wrong error: %v", err)
	}
	if err := setup(t).Insert("tests//l", "test"); err.Error() != "unwilling to operate on path with empty segment" {
		t.Errorf("wrong error: %v", err)
	}
	if err := setup(t).Insert("tests/", "test"); err.Error() != "path can NOT end with separator" {
		t.Errorf("wrong error: %v", err)
	}
	if err := setup(t).Insert("/tests", "test"); err.Error() != "path can NOT be rooted" {
		t.Errorf("wrong error: %v", err)
	}
	if err := setup(t).Insert("test", "test"); err.Error() != "input paths must contain at LEAST 2 components" {
		t.Errorf("wrong error: %v", err)
	}
	if err := setup(t).Insert("a", ""); err.Error() != "empty secret not allowed" {
		t.Errorf("wrong error: %v", err)
	}
	if err := setup(t).Insert(backend.NewPath("test", "offset", "value"), "pass"); err != nil {
		t.Errorf("no error: %v", err)
	}
	if err := fullSetup(t, true).Insert(backend.NewPath("test", "offset", "value"), "pass2"); err != nil {
		t.Errorf("wrong error: %v", err)
	}
	if err := fullSetup(t, true).Insert(backend.NewPath("test", "offset", "value2"), "pass\npass"); err != nil {
		t.Errorf("no error: %v", err)
	}
	q, err := fullSetup(t, true).Get(backend.NewPath("test", "offset", "value"), backend.SecretValue)
	if err != nil {
		t.Errorf("no error: %v", err)
	}
	if q.Value != "pass2" {
		t.Errorf("invalid retrieval")
	}
	q, err = fullSetup(t, true).Get(backend.NewPath("test", "offset", "value2"), backend.SecretValue)
	if err != nil {
		t.Errorf("no error: %v", err)
	}
	if q.Value != "pass\npass" {
		t.Errorf("invalid retrieval")
	}
	if err := fullSetup(t, true).Insert(backend.NewPath("test", "offset", "totp"), "5ae472abqdekjqykoyxk7hvc2leklq5n"); err != nil {
		t.Errorf("no error: %v", err)
	}
	if err := fullSetup(t, true).Insert(backend.NewPath("test", "offset", "totp"), "ljaf\n5ae472abqdekjqykoyxk7hvc2leklq5n"); err.Error() != "totp tokens can NOT be multi-line" {
		t.Errorf("wrong error: %v", err)
	}
}

func TestRemoves(t *testing.T) {
	if err := setup(t).Remove(nil); err.Error() != "entity is empty/invalid" {
		t.Errorf("wrong error: %v", err)
	}
	if err := setup(t).Remove(&backend.QueryEntity{}); err.Error() != "input paths must contain at LEAST 2 components" {
		t.Errorf("wrong error: %v", err)
	}
	if err := setup(t).Remove(&backend.QueryEntity{Path: backend.NewPath("test1", "test2", "test3")}); err.Error() != "failed to remove entity" {
		t.Errorf("wrong error: %v", err)
	}
	setup(t)
	for _, i := range []string{"test1", "test2"} {
		fullSetup(t, true).Insert(backend.NewPath(i, i, i), "pass")
	}
	if err := fullSetup(t, true).Remove(&backend.QueryEntity{Path: backend.NewPath("test1", "test1", "test1")}); err != nil {
		t.Errorf("wrong error: %v", err)
	}
	if err := check(t, backend.NewPath("test2", "test2", "test2")); err != nil {
		t.Errorf("invalid check: %v", err)
	}
	if err := fullSetup(t, true).Remove(&backend.QueryEntity{Path: backend.NewPath("test2", "test2", "test2")}); err != nil {
		t.Errorf("wrong error: %v", err)
	}
	setup(t)
	for _, i := range []string{backend.NewPath("test", "test", "test1"), backend.NewPath("test", "test", "test2"), backend.NewPath("test", "test", "test3"), backend.NewPath("test", "test1", "test2"), backend.NewPath("test", "test1", "test5")} {
		fullSetup(t, true).Insert(i, "pass")
	}
	if err := fullSetup(t, true).Remove(&backend.QueryEntity{Path: "test/test/test3"}); err != nil {
		t.Errorf("wrong error: %v", err)
	}
	if err := check(t, backend.NewPath("test", "test", "test2"), backend.NewPath("test", "test", "test1"), backend.NewPath("test", "test1", "test2"), backend.NewPath("test", "test1", "test5")); err != nil {
		t.Errorf("invalid check: %v", err)
	}
	if err := fullSetup(t, true).Remove(&backend.QueryEntity{Path: "test/test/test1"}); err != nil {
		t.Errorf("wrong error: %v", err)
	}
	if err := check(t, backend.NewPath("test", "test", "test2"), backend.NewPath("test", "test1", "test2"), backend.NewPath("test", "test1", "test5")); err != nil {
		t.Errorf("invalid check: %v", err)
	}
	if err := fullSetup(t, true).Remove(&backend.QueryEntity{Path: "test/test1/test5"}); err != nil {
		t.Errorf("wrong error: %v", err)
	}
	if err := check(t, backend.NewPath("test", "test", "test2"), backend.NewPath("test", "test1", "test2")); err != nil {
		t.Errorf("invalid check: %v", err)
	}
	if err := fullSetup(t, true).Remove(&backend.QueryEntity{Path: "test/test1/test2"}); err != nil {
		t.Errorf("wrong error: %v", err)
	}
	if err := check(t, backend.NewPath("test", "test", "test2")); err != nil {
		t.Errorf("invalid check: %v", err)
	}
	if err := fullSetup(t, true).Remove(&backend.QueryEntity{Path: "test/test/test2"}); err != nil {
		t.Errorf("wrong error: %v", err)
	}
}

func TestRemoveAlls(t *testing.T) {
	if err := setup(t).RemoveAll(nil); err.Error() != "no entities given" {
		t.Errorf("wrong error: %v", err)
	}
	if err := setup(t).RemoveAll([]backend.QueryEntity{}); err.Error() != "no entities given" {
		t.Errorf("wrong error: %v", err)
	}
	setup(t)
	for _, i := range []string{backend.NewPath("test", "test", "test1"), backend.NewPath("test", "test", "test2"), backend.NewPath("test", "test", "test3"), backend.NewPath("test", "test1", "test2"), backend.NewPath("test", "test1", "test5")} {
		fullSetup(t, true).Insert(i, "pass")
	}
	if err := fullSetup(t, true).RemoveAll([]backend.QueryEntity{{Path: "test/test/test3"}, {Path: "test/test/test1"}}); err != nil {
		t.Errorf("wrong error: %v", err)
	}
	if err := check(t, backend.NewPath("test", "test", "test2"), backend.NewPath("test", "test1", "test2"), backend.NewPath("test", "test1", "test5")); err != nil {
		t.Errorf("invalid check: %v", err)
	}
}

func check(t *testing.T, checks ...string) error {
	tr := fullSetup(t, true)
	for _, c := range checks {
		q, err := tr.Get(c, backend.BlankValue)
		if err != nil {
			return err
		}
		if q == nil {
			return fmt.Errorf("failed to find entity: %s", c)
		}
	}
	return nil
}

func TestKeyAndOrKeyFile(t *testing.T) {
	keyAndOrKeyFile(t, true, true)
	keyAndOrKeyFile(t, false, true)
	keyAndOrKeyFile(t, true, false)
	keyAndOrKeyFile(t, false, false)
}

func keyAndOrKeyFile(t *testing.T, key, keyFile bool) {
	os.Clearenv()
	file := testFile("keyorkeyfile.kdbx")
	os.Remove(file)
	os.Setenv("LOCKBOX_STORE", file)
	if key {
		os.Setenv("LOCKBOX_KEY", "test")
		os.Setenv("LOCKBOX_KEYMODE", "plaintext")
	} else {
		os.Setenv("LOCKBOX_KEYMODE", "none")
	}
	if keyFile {
		key := testFile("keyfileor.key")
		os.Setenv("LOCKBOX_KEYFILE", key)
		os.WriteFile(key, []byte("test"), 0o644)
	}
	tr, err := backend.NewTransaction()
	if err != nil {
		t.Errorf("failed: %v", err)
	}
	invalid := !key && !keyFile
	err = tr.Insert(backend.NewPath("a", "b"), "t")
	if invalid {
		if err == nil || err.Error() != "key and/or keyfile must be set" {
			t.Errorf("invalid error: %v", err)
		}
	} else {
		if err != nil {
			t.Errorf("no error allowed: %v", err)
		}
	}
}

func TestReKey(t *testing.T) {
	os.Clearenv()
	f := "rekey_test.kdbx"
	file := testFile(f)
	defer os.Remove(filepath.Join(testDir, f))
	os.Setenv("LOCKBOX_READONLY", "no")
	os.Setenv("LOCKBOX_STORE", file)
	os.Setenv("LOCKBOX_KEY", "test")
	os.Setenv("LOCKBOX_KEYMODE", "plaintext")
	os.Setenv("LOCKBOX_TOTP", "totp")
	os.Setenv("LOCKBOX_HOOKDIR", "")
	os.Setenv("LOCKBOX_SET_MODTIME", "")
	tr, err := backend.NewTransaction()
	if err != nil {
		t.Errorf("failed: %v", err)
	}
	if err := tr.ReKey("", ""); err == nil || err.Error() != "key and/or keyfile must be set" {
		t.Errorf("no error: %v", err)
	}
	if err := tr.ReKey("abc", ""); err != nil {
		t.Errorf("no error: %v", err)
	}
}
