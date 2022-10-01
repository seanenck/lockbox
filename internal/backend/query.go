// Package backend handles querying a store
package backend

import (
	"crypto/sha512"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Get will request a singular entity
func (t *Transaction) Get(path string, mode ValueMode) (*QueryEntity, error) {
	e, err := t.QueryCallback(QueryOptions{Mode: ExactMode, Criteria: path, Values: mode})
	if err != nil {
		return nil, err
	}
	switch len(e) {
	case 0:
		return nil, nil
	case 1:
		return &e[0], nil
	default:
		return nil, errors.New("too many entities matched")
	}
}

// QueryCallback will retrieve a query based on setting
func (t *Transaction) QueryCallback(args QueryOptions) ([]QueryEntity, error) {
	if args.Mode == noneMode {
		return nil, errors.New("no query mode specified")
	}
	var keys []string
	entities := make(map[string]QueryEntity)
	isSort := args.Mode == ListMode || args.Mode == FindMode || args.Mode == SuffixMode
	decrypt := args.Values != BlankValue
	err := t.act(func(ctx Context) error {
		for _, entry := range ctx.db.Content.Root.Groups[0].Entries {
			path := getPathName(entry)
			if isSort {
				switch args.Mode {
				case FindMode:
					if !strings.Contains(path, args.Criteria) {
						continue
					}
				case SuffixMode:
					if !strings.HasSuffix(path, args.Criteria) {
						continue
					}
				}

			} else {
				if args.Mode == ExactMode {
					if path != args.Criteria {
						continue
					}
				}
			}
			keys = append(keys, path)
			entities[path] = QueryEntity{backing: entry}
		}
		if decrypt {
			return ctx.db.UnlockProtectedEntries()
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if isSort {
		sort.Strings(keys)
	}
	var results []QueryEntity
	for _, k := range keys {
		entity := QueryEntity{Path: k}
		if args.Values != BlankValue {
			e := entities[k]
			val := getValue(e.backing, notesKey)
			if strings.TrimSpace(val) == "" {
				val = e.backing.GetPassword()
			}
			switch args.Values {
			case SecretValue:
				entity.Value = val
			case HashedValue:
				entity.Value = fmt.Sprintf("%x", sha512.Sum512([]byte(val)))
			}
		}
		results = append(results, entity)
	}
	return results, nil
}

// NewSuffix creates a new user 'name' suffix
func NewSuffix(name string) string {
	return fmt.Sprintf("%c%s", os.PathSeparator, name)
}

// NewPath creates a new storage location path.
func NewPath(segments ...string) string {
	return filepath.Join(segments...)
}

// Directory gets the offset location of the entry without the 'name'
func (e QueryEntity) Directory() string {
	return filepath.Dir(e.Path)
}