// Package completions generations shell completions
package completions

import (
	"bytes"
	"embed"
	"fmt"
	"slices"
	"sort"
	"text/template"

	"github.com/seanenck/lockbox/internal/app/commands"
	"github.com/seanenck/lockbox/internal/config"
	"github.com/seanenck/lockbox/internal/util"
)

type (
	// Template handles the inputs to completions for templating
	Template struct {
		InsertCommand       string
		TOTPListCommand     string
		RemoveCommand       string
		ClipCommand         string
		ShowCommand         string
		MultiLineCommand    string
		MoveCommand         string
		TOTPCommand         string
		DoTOTPList          string
		DoList              string
		Executable          string
		JSONCommand         string
		HelpCommand         string
		HelpAdvancedCommand string
		HelpConfigCommand   string
		ExportCommand       string
		Options             []CompletionOption
		TOTPSubCommands     []CompletionOption
		Conditionals        Conditionals
	}
	// Conditionals help control completion flow
	Conditionals struct {
		Not struct {
			ReadOnly       string
			CanClip        string
			CanTOTP        string
			AskMode        string
			Ever           string
			CanPasswordGen string
		}
		Exported []string
	}
	// CompletionOption are conditional wrapped logic for options that may be disabled
	CompletionOption struct {
		Conditional string
		Key         string
	}
)

//go:embed shell/*
var shell embed.FS

func (c Template) newGenOptions(defaults []string, kv map[string]string) []CompletionOption {
	opt := []CompletionOption{}
	for _, a := range defaults {
		opt = append(opt, CompletionOption{c.Conditionals.Not.Ever, a})
	}
	var keys []string
	for k := range kv {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, key := range keys {
		check := kv[key]
		opt = append(opt, CompletionOption{check, key})
	}
	return opt
}

// NewConditionals creates the conditional components of completions
func NewConditionals() Conditionals {
	const shellIsNotText = `[ "%s" != "%s" ]`
	c := Conditionals{}
	registerIsNotEqual := func(key interface{ Key() string }, right string) string {
		k := key.Key()
		c.Exported = append(c.Exported, k)
		return fmt.Sprintf(shellIsNotText, fmt.Sprintf("$%s", k), right)
	}
	c.Not.ReadOnly = registerIsNotEqual(config.EnvReadOnly, config.YesValue)
	c.Not.CanClip = registerIsNotEqual(config.EnvClipEnabled, config.NoValue)
	c.Not.CanTOTP = registerIsNotEqual(config.EnvTOTPEnabled, config.NoValue)
	c.Not.AskMode = registerIsNotEqual(config.EnvPasswordMode, string(config.AskKeyMode))
	c.Not.CanPasswordGen = registerIsNotEqual(config.EnvPasswordGenEnabled, config.NoValue)
	c.Not.Ever = fmt.Sprintf(shellIsNotText, "1", "0")
	return c
}

// Generate handles creating shell completion outputs
func Generate(completionType, exe string) ([]string, error) {
	if !slices.Contains(commands.CompletionTypes, completionType) {
		return nil, fmt.Errorf("unknown completion request: %s", completionType)
	}
	c := Template{
		Executable:          exe,
		InsertCommand:       commands.Insert,
		RemoveCommand:       commands.Remove,
		TOTPListCommand:     commands.TOTPList,
		ClipCommand:         commands.Clip,
		ShowCommand:         commands.Show,
		MultiLineCommand:    commands.MultiLine,
		JSONCommand:         commands.JSON,
		HelpCommand:         commands.Help,
		HelpAdvancedCommand: commands.HelpAdvanced,
		HelpConfigCommand:   commands.HelpConfig,
		TOTPCommand:         commands.TOTP,
		MoveCommand:         commands.Move,
		DoList:              fmt.Sprintf("%s %s", exe, commands.List),
		DoTOTPList:          fmt.Sprintf("%s %s %s", exe, commands.TOTP, commands.TOTPList),
		ExportCommand:       fmt.Sprintf("%s %s %s", exe, commands.Env, commands.Completions),
	}
	c.Conditionals = NewConditionals()

	c.Options = c.newGenOptions([]string{commands.Env, commands.Help, commands.List, commands.Show, commands.Version, commands.JSON},
		map[string]string{
			commands.Clip:             c.Conditionals.Not.CanClip,
			commands.TOTP:             c.Conditionals.Not.CanTOTP,
			commands.Move:             c.Conditionals.Not.ReadOnly,
			commands.Remove:           c.Conditionals.Not.ReadOnly,
			commands.Insert:           c.Conditionals.Not.ReadOnly,
			commands.MultiLine:        c.Conditionals.Not.ReadOnly,
			commands.PasswordGenerate: c.Conditionals.Not.CanPasswordGen,
		})
	c.TOTPSubCommands = c.newGenOptions([]string{commands.TOTPMinimal, commands.TOTPOnce, commands.TOTPShow},
		map[string]string{
			commands.TOTPClip:   c.Conditionals.Not.CanClip,
			commands.TOTPInsert: c.Conditionals.Not.ReadOnly,
		})
	using, err := util.ReadDirFile("shell", fmt.Sprintf("%s.sh", completionType), shell)
	if err != nil {
		return nil, err
	}
	s, err := templateScript(using, c)
	if err != nil {
		return nil, err
	}
	return []string{s}, nil
}

func templateScript(script string, c Template) (string, error) {
	t, err := template.New("t").Parse(script)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, c); err != nil {
		return "", err
	}
	return buf.String(), nil
}
