// Package config handles user inputs/UI elements.
package config

import (
	"fmt"
	"slices"
	"strings"

	"github.com/seanenck/lockbox/internal/config/store"
)

type (
	environmentBase struct {
		key         string
		description string
		requirement string
	}
	environmentDefault[T any] struct {
		environmentBase
		value T
	}
	// EnvironmentInt are environment settings that are integers
	EnvironmentInt struct {
		environmentDefault[int]
		canZero bool
		short   string
	}
	// EnvironmentBool are environment settings that are booleans
	EnvironmentBool struct {
		environmentDefault[bool]
	}
	// EnvironmentStrings are string-based settings
	EnvironmentStrings struct {
		environmentDefault[string]
		flags   []stringsFlags
		allowed []string
	}
	// EnvironmentFormatter allows for sending a string into a get request
	EnvironmentFormatter struct {
		environmentBase
		allowed string
		fxn     func(string, string) string
	}
	metaData struct {
		value     string
		allowed   []string
		tomlType  tomlType
		tomlValue string
		canExpand bool
	}
)

func (e environmentBase) Key() string {
	return fmt.Sprintf(environmentPrefix+"%s", e.key)
}

// Get will get the boolean value for the setting
func (e EnvironmentBool) Get() bool {
	val, ok := store.GetBool(e.Key())
	if !ok {
		val = e.value
	}
	return val
}

// Get will get the integer value for the setting
func (e EnvironmentInt) Get() (int64, error) {
	i, ok := store.GetInt64(e.Key())
	if !ok {
		i = int64(e.value)
	}
	invalid := false
	check := ""
	if e.canZero {
		check = "="
	}
	switch i {
	case 0:
		invalid = !e.canZero
	default:
		invalid = i < 0
	}
	if invalid {
		return -1, fmt.Errorf("%s must be >%s 0", e.short, check)
	}
	return i, nil
}

// Get will read the string from the environment
func (e EnvironmentStrings) Get() string {
	return stringsGet(e, store.GetString, func(val string) string {
		return val
	})
}

// AsArray indicates the item should be queried as an array
func (e EnvironmentStrings) AsArray() []string {
	return stringsGet(e, store.GetArray, func(val string) []string {
		return strings.Split(val, arrayDelimiter)
	})
}

func stringsGet[T string | []string](e EnvironmentStrings, getter func(string) (T, bool), conv func(string) T) T {
	val, ok := getter(e.Key())
	if !ok {
		flags := e.flattenFlags()
		if slices.Contains(flags, canDefaultFlag) {
			val = conv(e.value)
		}
	}
	return val
}

// Get will retrieve the value with the formatted input included
func (e EnvironmentFormatter) Get(value string) string {
	return e.fxn(e.Key(), value)
}

func (e EnvironmentStrings) display() metaData {
	var t tomlType
	t = tomlString
	v := "\"\""
	show := e.allowed
	value := e.value
	flags := e.flattenFlags()
	canExpand := slices.Contains(flags, canExpandFlag)
	if slices.Contains(flags, isArrayFlag) {
		t = tomlArray
		v = "[]"
		if slices.Contains(flags, isCommandFlag) {
			if len(show) == 0 {
				show = []string{"[cmd args...]"}
			}
			if value == "" {
				value = "(detected)"
			}
		}
	}
	return metaData{
		value:     value,
		allowed:   show,
		tomlType:  t,
		tomlValue: v,
		canExpand: canExpand,
	}
}

func (e environmentBase) self() environmentBase {
	return e
}

func (e EnvironmentBool) display() metaData {
	val := NoValue
	if e.value {
		val = YesValue
	}
	return metaData{
		value:     val,
		allowed:   []string{YesValue, NoValue},
		tomlType:  tomlBool,
		tomlValue: YesValue,
		canExpand: false,
	}
}

func (e EnvironmentInt) display() metaData {
	return metaData{
		value:     fmt.Sprintf("%d", e.value),
		allowed:   []string{"<integer>"},
		tomlType:  tomlInt,
		tomlValue: "0",
		canExpand: false,
	}
}

func (e EnvironmentFormatter) display() metaData {
	return metaData{
		value:     strings.ReplaceAll(strings.ReplaceAll(EnvTOTPFormat.Get("%s"), "%25s", "%s"), "&", " \\\n           &"),
		allowed:   []string{e.allowed},
		tomlType:  tomlString,
		tomlValue: "\"\"",
		canExpand: false,
	}
}

func (e EnvironmentStrings) flattenFlags() []stringsFlags {
	flags := e.flags
	if slices.Contains(e.flags, isCommandFlag) {
		flags = append(flags, canExpandFlag, isArrayFlag)
	}
	return flags
}
