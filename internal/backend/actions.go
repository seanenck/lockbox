// Package backend handles kdbx interactions
package backend

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/enckse/lockbox/internal/inputs"
	"github.com/tobischo/gokeepasslib/v3"
	"github.com/tobischo/gokeepasslib/v3/wrappers"
)

func (t *Transaction) act(cb action) error {
	if !t.valid {
		return errors.New("invalid transaction")
	}
	key, err := inputs.GetKey("", "")
	if err != nil {
		return err
	}
	k := string(key)
	if !t.exists {
		if err := create(t.file, k); err != nil {
			return err
		}
	}
	f, err := os.Open(t.file)
	if err != nil {
		return err
	}
	defer f.Close()
	db := gokeepasslib.NewDatabase()
	db.Credentials = gokeepasslib.NewPasswordCredentials(k)
	if err := gokeepasslib.NewDecoder(f).Decode(db); err != nil {
		return err
	}
	if len(db.Content.Root.Groups) != 1 {
		return errors.New("kdbx must only have ONE root group")
	}
	cErr := cb(Context{db: db})
	if t.write {
		if err := db.LockProtectedEntries(); err != nil {
			return err
		}
		if err := f.Close(); err != nil {
			return err
		}
		f, err = os.Create(t.file)
		if err != nil {
			return err
		}
		defer f.Close()
		return encode(f, db)
	}
	return cErr
}

func (t *Transaction) change(cb action) error {
	return t.act(func(c Context) error {
		if err := c.db.UnlockProtectedEntries(); err != nil {
			return err
		}
		t.write = true
		return cb(c)
	})
}

// Insert handles inserting a new element
func (t *Transaction) Insert(path, val string, entity *QueryEntity, multi bool) error {
	return t.change(func(c Context) error {
		if entity != nil {
			if err := remove(entity, c); err != nil {
				return err
			}
		}
		e := gokeepasslib.NewEntry()
		e.Values = append(e.Values, value(titleKey, filepath.Dir(path)))
		e.Values = append(e.Values, value(userNameKey, filepath.Base(path)))
		field := passKey
		if multi {
			field = notesKey
		}

		e.Values = append(e.Values, protectedValue(field, val))
		c.db.Content.Root.Groups[0].Entries = append(c.db.Content.Root.Groups[0].Entries, e)
		return nil
	})
}

func remove(entity *QueryEntity, c Context) error {
	entries := c.db.Content.Root.Groups[0].Entries
	idx := -1
	for i, e := range entries {
		if entity.Path == getPathName(e) {
			idx = i
		}
	}
	if idx < 0 {
		return errors.New("unable to select entity for deletion")
	}
	c.db.Content.Root.Groups[0].Entries = append(entries[:idx], entries[idx+1:]...)
	return nil
}

// Remove handles remove an element
func (t *Transaction) Remove(entity *QueryEntity) error {
	if entity == nil {
		return errors.New("entity is empty/invalid")
	}
	return t.change(func(c Context) error {
		return remove(entity, c)
	})
}

func getValue(e gokeepasslib.Entry, key string) string {
	v := e.Get(key)
	if v == nil {
		return ""
	}
	return v.Value.Content
}

func value(key string, value string) gokeepasslib.ValueData {
	return gokeepasslib.ValueData{Key: key, Value: gokeepasslib.V{Content: value}}
}

func getPathName(entry gokeepasslib.Entry) string {
	return filepath.Join(entry.GetTitle(), getValue(entry, userNameKey))
}

func protectedValue(key string, value string) gokeepasslib.ValueData {
	return gokeepasslib.ValueData{
		Key:   key,
		Value: gokeepasslib.V{Content: value, Protected: wrappers.NewBoolWrapper(true)},
	}
}