package config

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"

	"github.com/BurntSushi/toml"
)

const isInclude = "include"

type (
	// Loader indicates how included files should be sourced
	Loader func(string) (io.Reader, error)
	// ShellEnv is the output shell environment settings parsed from TOML config
	ShellEnv struct {
		Key   string
		Value string
		raw   string
	}
)

var (
	// ExampleTOML is an example TOML file of viable fields
	//go:embed "config.toml"
	ExampleTOML string
	redirects   = map[string]string{
		"HOOKS_DIRECTORY":           EnvHookDir.Key(),
		"HOOKS_ENABLED":             EnvNoHooks.Key(),
		"JSON_MODE":                 EnvJSONDataOutput.Key(),
		"JSON_HASH_LENGTH":          EnvHashLength.Key(),
		"CREDENTIALS_KEY_FILE":      EnvKeyFile.Key(),
		"CREDENTIALS_PASSWORD_MODE": EnvKeyMode.Key(),
		"CREDENTIALS_PASSWORD":      envKey.Key(),
		"CLIP_ENABLED":              EnvNoClip.Key(),
		"COLOR_ENABLED":             EnvNoColor.Key(),
		"PWGEN_ENABLED":             EnvNoPasswordGen.Key(),
		"TOTP_ENABLED":              EnvNoTOTP.Key(),
		"TOTP_ATTRIBUTE":            EnvTOTPToken.Key(),
		"TOTP_OTP_FORMAT":           EnvFormatTOTP.Key(),
		"TOTP_COLOR_WINDOWS":        EnvTOTPColorBetween.Key(),
		"TOTP_TIMEOUT":              EnvMaxTOTP.Key(),
		"DEFAULTS_MODTIME":          EnvModTime.Key(),
		"DEFAULTS_COMPLETION":       EnvDefaultCompletion.Key(),
		"PWGEN_WORDS_COMMAND":       EnvPasswordGenWordList.Key(),
		"CLIP_COPY_COMMAND":         EnvClipCopy.Key(),
		"CLIP_PASTE_COMMAND":        EnvClipPaste.Key(),
		"CLIP_TIMEOUT":              EnvClipMax.Key(),
		"PWGEN_CHARACTERS":          EnvPasswordGenChars.Key(),
	}
	arrayTypes = []string{
		EnvClipCopy.Key(),
		EnvClipPaste.Key(),
		EnvPasswordGenWordList.Key(),
		envKey.Key(),
		EnvTOTPColorBetween.Key(),
	}
	intTypes = []string{
		EnvClipMax.Key(),
		EnvMaxTOTP.Key(),
		EnvHashLength.Key(),
		EnvPasswordGenCount.Key(),
	}
	boolTypes = []string{
		EnvClipOSC52.Key(),
		EnvNoClip.Key(),
		EnvNoColor.Key(),
		EnvNoHooks.Key(),
		EnvNoPasswordGen.Key(),
		EnvNoTOTP.Key(),
		EnvPasswordGenTitle.Key(),
		EnvReadOnly.Key(),
		EnvInteractive.Key(),
	}
	reverseMap = map[string][]string{
		"[]":   arrayTypes,
		"0":    intTypes,
		"true": boolTypes,
	}
)

// DefaultTOML will load the internal, default TOML with additional comment markups
func DefaultTOML() (string, error) {
	s, err := LoadConfig(strings.NewReader(ExampleTOML), nil)
	if err != nil {
		return "", err
	}
	const root = "_root_"
	unmapped := make(map[string][]string)
	keys := []string{}
	for _, item := range s {
		raw := item.raw
		parts := strings.Split(raw, "_")
		length := len(parts)
		if length == 0 {
			return "", fmt.Errorf("invalid internal TOML structure: %v", item)
		}
		key := parts[0]
		sub := ""
		switch length {
		case 1:
			key = root
			sub = parts[0]
		case 2:
			sub = parts[1]
		default:
			sub = strings.Join(parts[1:], "_")
		}
		field := "\"\""
		for to, fromKey := range reverseMap {
			if slices.Contains(fromKey, item.Key) {
				field = to
				break
			}
		}
		text, err := generateDetailText(item.Key)
		if err != nil {
			return "", err
		}
		sub = fmt.Sprintf(`%s
%s = %s
`, text, sub, field)
		had, ok := unmapped[key]
		if !ok {
			had = []string{}
			keys = append(keys, key)
		}
		had = append(had, sub)
		unmapped[key] = had
	}
	sort.Strings(keys)
	builder := strings.Builder{}
	configEnv, err := generateDetailText(EnvConfig.Key())
	if err != nil {
		return "", err
	}
	for _, header := range []string{configEnv, "\n", fmt.Sprintf(`
# include additional configs, can NOT nest, but does allow globs ('*')
# this field is not configurable via environment variables
# and it is not considered part of the environment either
# it is ONLY used during TOML configuration loading
%s = []
`, isInclude), "\n"} {
		if _, err := builder.WriteString(header); err != nil {
			return "", err
		}
	}
	for _, k := range keys {
		if k != root {
			if _, err := fmt.Fprintf(&builder, "\n[%s]\n", k); err != nil {
				return "", err
			}
		}
		for _, sub := range unmapped[k] {
			if _, err := builder.WriteString(sub); err != nil {
				return "", err
			}
		}
	}
	return builder.String(), nil
}

func generateDetailText(key string) (string, error) {
	data, ok := registry[key]
	if !ok {
		return "", fmt.Errorf("unexpected configuration key has no environment settings: %s", key)
	}
	env := data.self()
	value, allow := data.values()
	if len(value) == 0 {
		value = "(unset)"
	}
	description := strings.TrimSpace(Wrap(2, env.desc))
	requirement := "optional/default"
	r := strings.TrimSpace(env.requirement)
	if r != "" {
		requirement = r
	}
	var text []string
	for _, line := range []string{fmt.Sprintf("environment: %s", key), fmt.Sprintf("description:\n%s\n", description), fmt.Sprintf("default: %s", requirement), fmt.Sprintf("option: %s", strings.Join(allow, "|"))} {
		for _, comment := range strings.Split(line, "\n") {
			text = append(text, fmt.Sprintf("# %s", comment))
		}
	}
	return strings.Join(text, "\n"), nil
}

// LoadConfig will read the input reader and use the loader to source configuration files
func LoadConfig(r io.Reader, loader Loader) ([]ShellEnv, error) {
	m := make(map[string]interface{})
	if err := overlayConfig(r, true, &m, loader); err != nil {
		return nil, err
	}
	m = flatten(m, "")
	var res []ShellEnv
	for k, v := range m {
		export := strings.ToUpper(k)
		if r, ok := redirects[export]; ok {
			export = r
		} else {
			export = environmentPrefix + export
		}
		if _, ok := registry[export]; !ok {
			return nil, fmt.Errorf("unknown key: %s (%s)", k, export)
		}
		value, ok := v.(string)
		if !ok {
			if slices.Contains(arrayTypes, export) {
				array, err := parseStringArray(v)
				if err != nil {
					return nil, err
				}
				value = strings.Join(array, " ")
			} else if slices.Contains(intTypes, export) {
				i, ok := v.(int64)
				if !ok {
					return nil, fmt.Errorf("non-int64 found where expected: %v", v)
				}
				if i < 0 {
					return nil, fmt.Errorf("%d is negative (not allowed here)", i)
				}
				value = fmt.Sprintf("%d", i)
			} else if slices.Contains(boolTypes, export) {
				switch t := v.(type) {
				case bool:
					if t {
						value = yes
					} else {
						value = no
					}
				default:
					return nil, fmt.Errorf("non-bool found where expected: %v", v)
				}
			} else {
				return nil, fmt.Errorf("unknown field, can't determine type: %s (%v)", k, v)
			}
		}
		value = os.Expand(value, os.Getenv)
		res = append(res, ShellEnv{Key: export, Value: value, raw: k})

	}
	return res, nil
}

func overlayConfig(r io.Reader, canInclude bool, m *map[string]interface{}, loader Loader) error {
	d := toml.NewDecoder(r)
	if _, err := d.Decode(m); err != nil {
		return err
	}
	res := *m
	includes, ok := res[isInclude]
	if ok {
		delete(*m, isInclude)
		including, err := parseStringArray(includes)
		if err != nil {
			return err
		}
		if len(including) > 0 {
			if !canInclude {
				return errors.New("nested includes not allowed")
			}
			for _, s := range including {
				use := os.Expand(s, os.Getenv)
				files := []string{use}
				if strings.Contains(use, "*") {
					matched, err := filepath.Glob(use)
					if err != nil {
						return err
					}
					files = matched
				}
				for _, file := range files {
					reader, err := loader(file)
					if err != nil {
						return err
					}
					if err := overlayConfig(reader, false, m, nil); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

func parseStringArray(value interface{}) ([]string, error) {
	var res []string
	switch t := value.(type) {
	case []interface{}:
		for _, item := range t {
			switch s := item.(type) {
			case string:
				res = append(res, s)
			default:
				return nil, fmt.Errorf("value is not string in array: %v", item)
			}
		}
	default:
		return nil, fmt.Errorf("value is not of array type: %v", value)
	}
	return res, nil
}

func flatten(m map[string]interface{}, prefix string) map[string]interface{} {
	flattened := make(map[string]interface{})

	for k, v := range m {
		key := k
		if prefix != "" {
			key = prefix + "_" + k
		}

		switch to := v.(type) {
		case map[string]interface{}:
			for subKey, subVal := range flatten(to, key) {
				flattened[subKey] = subVal
			}
		default:
			flattened[key] = v
		}
	}

	return flattened
}

func configLoader(path string) (io.Reader, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(data), nil
}

// LoadConfigFile will load a path as the configuration
// it will also set the environment
func LoadConfigFile(path string) error {
	reader, err := configLoader(path)
	if err != nil {
		return err
	}
	env, err := LoadConfig(reader, configLoader)
	if err != nil {
		return err
	}
	for _, v := range env {
		os.Setenv(v.Key, v.Value)
	}
	return nil
}