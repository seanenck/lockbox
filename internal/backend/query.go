// Package backend handles querying a store
package backend

import (
	"crypto/sha512"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/seanenck/lockbox/internal/config"
	"github.com/seanenck/lockbox/internal/output"
	"github.com/tobischo/gokeepasslib/v3"
)

type (
	// QueryOptions indicates how to find entities
	QueryOptions struct {
		Mode     QueryMode
		Values   ValueMode
		Criteria string
	}
	// JSON is an entry as a JSON string
	JSON struct {
		ModTime string `json:"modtime"`
		Data    string `json:"data,omitempty"`
	}
	// QueryMode indicates HOW an entity will be found
	QueryMode int
	// ValueMode indicates what to do with the store value of the entity
	ValueMode int
)

const (
	// BlankValue will not decrypt secrets, empty value
	BlankValue ValueMode = iota
	// SecretValue will have the raw secret onboard
	SecretValue
	// JSONValue will show entries as a JSON payload
	JSONValue
)

const (
	noneMode QueryMode = iota
	// ListMode indicates ALL entities will be listed
	ListMode
	// FindMode indicates a _contains_ search for an entity
	FindMode
	// ExactMode means an entity must MATCH the string exactly
	ExactMode
	// SuffixMode will look for an entity ending in a specific value
	SuffixMode
	// PrefixMode allows for entities starting with a specific value
	PrefixMode
)

// MatchPath will try to match 1 or more elements (more elements when globbing)
func (t *Transaction) MatchPath(path string) ([]Entity, error) {
	if !strings.HasSuffix(path, isGlob) {
		e, err := t.Get(path, BlankValue)
		if err != nil {
			return nil, err
		}
		if e == nil {
			return nil, nil
		}
		return []Entity{*e}, nil
	}
	prefix := strings.TrimSuffix(path, isGlob)
	if strings.HasSuffix(prefix, pathSep) {
		return nil, errors.New("invalid match criteria, too many path separators")
	}
	return t.queryCollect(QueryOptions{Mode: PrefixMode, Criteria: prefix + pathSep, Values: BlankValue})
}

// Get will request a singular entity
func (t *Transaction) Get(path string, mode ValueMode) (*Entity, error) {
	_, _, err := splitComponents(path)
	if err != nil {
		return nil, err
	}
	e, err := t.queryCollect(QueryOptions{Mode: ExactMode, Criteria: path, Values: mode})
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

func forEach(offset string, groups []gokeepasslib.Group, entries []gokeepasslib.Entry, cb func(string, gokeepasslib.Entry)) {
	for _, g := range groups {
		o := ""
		if offset == "" {
			o = g.Name
		} else {
			o = NewPath(offset, g.Name)
		}
		forEach(o, g.Groups, g.Entries, cb)
	}
	for _, e := range entries {
		cb(offset, e)
	}
}

func (t *Transaction) queryCollect(args QueryOptions) ([]Entity, error) {
	e, err := t.QueryCallback(args)
	if err != nil {
		return nil, err
	}
	return e.Collect()
}

// QueryCallback will retrieve a query based on setting
func (t *Transaction) QueryCallback(args QueryOptions) (QuerySeq2, error) {
	if args.Mode == noneMode {
		return nil, errors.New("no query mode specified")
	}
	type entity struct {
		path    string
		backing gokeepasslib.Entry
	}
	var entities []entity
	isSort := args.Mode != ExactMode
	decrypt := args.Values != BlankValue
	err := t.act(func(ctx Context) error {
		forEach("", ctx.db.Content.Root.Groups[0].Groups, ctx.db.Content.Root.Groups[0].Entries, func(offset string, entry gokeepasslib.Entry) {
			path := getPathName(entry)
			if offset != "" {
				path = NewPath(offset, path)
			}
			if isSort {
				switch args.Mode {
				case FindMode:
					if !strings.Contains(path, args.Criteria) {
						return
					}
				case SuffixMode:
					if !strings.HasSuffix(path, args.Criteria) {
						return
					}
				case PrefixMode:
					if !strings.HasPrefix(path, args.Criteria) {
						return
					}
				}
			} else {
				if args.Mode == ExactMode {
					if path != args.Criteria {
						return
					}
				}
			}
			obj := entity{backing: entry, path: path}
			if isSort && len(entities) > 0 {
				i, _ := slices.BinarySearchFunc(entities, obj, func(i, j entity) int {
					return strings.Compare(i.path, j.path)
				})
				entities = slices.Insert(entities, i, obj)
			} else {
				entities = append(entities, obj)
			}
		})
		if decrypt {
			return ctx.db.UnlockProtectedEntries()
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	jsonMode := output.JSONModes.Blank
	if args.Values == JSONValue {
		m, err := output.ParseJSONMode(config.EnvJSONMode.Get())
		if err != nil {
			return nil, err
		}
		jsonMode = m
	}
	var hashLength int64
	if jsonMode == output.JSONModes.Hash {
		hashLength, err = config.EnvJSONHashLength.Get()
		if err != nil {
			return nil, err
		}
	}
	l := int(hashLength)
	return func(yield func(Entity, error) bool) {
		for _, item := range entities {
			entity := Entity{Path: item.path}
			var err error
			if args.Values != BlankValue {
				val := getValue(item.backing, notesKey)
				if strings.TrimSpace(val) == "" {
					val = item.backing.GetPassword()
				}
				switch args.Values {
				case JSONValue:
					data := ""
					switch jsonMode {
					case output.JSONModes.Raw:
						data = val
					case output.JSONModes.Hash:
						data = fmt.Sprintf("%x", sha512.Sum512([]byte(val)))
						if hashLength > 0 && len(data) > l {
							data = data[0:hashLength]
						}
					}
					t := getValue(item.backing, modTimeKey)
					s := JSON{ModTime: t, Data: data}
					m, jErr := json.Marshal(s)
					if jErr == nil {
						entity.Value = string(m)
					} else {
						err = jErr
					}
				case SecretValue:
					entity.Value = val
				}
			}
			if !yield(entity, err) {
				return
			}
		}
	}, nil
}
