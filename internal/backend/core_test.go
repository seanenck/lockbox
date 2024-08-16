package backend_test

import (
	"fmt"
	"testing"

	"github.com/seanenck/lockbox/internal/backend"
)

func TestLoad(t *testing.T) {
	if _, err := backend.Load("  "); err.Error() != "no store set" {
		t.Errorf("invalid error: %v", err)
	}
	if _, err := backend.Load("garbage"); err.Error() != "should use a .kdbx extension" {
		t.Errorf("invalid error: %v", err)
	}
	if _, err := backend.Load("garbage.kdbx"); err.Error() != "invalid file, does not exist" {
		t.Errorf("invalid error: %v", err)
	}
}

func TestIsDirectory(t *testing.T) {
	if backend.IsDirectory("") {
		t.Error("invalid directory detection")
	}
	if !backend.IsDirectory("/") {
		t.Error("invalid directory detection")
	}
	if backend.IsDirectory("/a") {
		t.Error("invalid directory detection")
	}
}

func TestQueryToTransaction(t *testing.T) {
	q := backend.QueryEntity{Path: "abc", Value: "xyz"}
	tx := q.Transaction()
	if fmt.Sprintf("%v", tx) != "{abc xyz}" {
		t.Errorf("invalid transaction: %v", tx)
	}
}

func TestBase(t *testing.T) {
	b := backend.Base("")
	if b != "" {
		t.Error("invalid base")
	}
	b = backend.Base("aaa")
	if b != "aaa" {
		t.Error("invalid base")
	}
	b = backend.Base("aaa/")
	if b != "" {
		t.Error("invalid base")
	}
	b = backend.Base("aaa/a")
	if b != "a" {
		t.Error("invalid base")
	}
}

func TestDirectory(t *testing.T) {
	b := backend.Directory("")
	if b != "" {
		t.Error("invalid directory")
	}
	b = backend.Directory("/")
	if b != "" {
		t.Error("invalid directory")
	}
	b = backend.Directory("/a")
	if b != "" {
		t.Error("invalid directory")
	}
	b = backend.Directory("a")
	if b != "" {
		t.Error("invalid directory")
	}
	b = backend.Directory("b/a")
	if b != "b" {
		t.Error("invalid directory")
	}
}

func TestEntityDir(t *testing.T) {
	q := backend.QueryEntity{Path: backend.NewPath("abc", "xyz")}
	if q.Directory() != "abc" {
		t.Error("invalid query directory")
	}
	q = backend.QueryEntity{Path: backend.NewPath("abc", "xyz", "111")}
	if q.Directory() != "abc/xyz" {
		t.Error("invalid query directory")
	}
	q = backend.QueryEntity{Path: ""}
	if q.Directory() != "" {
		t.Error("invalid query directory")
	}
	q = backend.QueryEntity{Path: backend.NewPath("abc")}
	if q.Directory() != "" {
		t.Error("invalid query directory")
	}
}

func TestNewPath(t *testing.T) {
	p := backend.NewPath("abc", "xyz")
	if p != backend.NewPath("abc", "xyz") {
		t.Error("invalid new path")
	}
}

func TestNewSuffix(t *testing.T) {
	if backend.NewSuffix("test") != "/test" {
		t.Error("invalid suffix")
	}
}
