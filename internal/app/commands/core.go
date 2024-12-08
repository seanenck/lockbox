// Package commands defines available commands within the app
package commands

const (
	// TOTP is the parent of totp and by defaults generates a rotating token
	TOTP = "totp"
	// Conv handles text conversion of the data store
	Conv = "conv"
	// Clear is a callback to manage clipboard clearing
	Clear = "clear"
	// Clip will copy values to the clipboard
	Clip = "clip"
	// Find is for simplistic searching of entries
	Find = "find"
	// Insert adds a value
	Insert = "insert"
	// List lists all entries
	List = "ls"
	// Move will move source to destination
	Move = "mv"
	// Show will show the value in an entry
	Show = "show"
	// Version displays version information
	Version = "version"
	// Help shows usage
	Help = "help"
	// HelpAdvanced shows advanced help
	HelpAdvanced = "verbose"
	// HelpConfig shows configuration information
	HelpConfig = "config"
	// Remove removes an entry
	Remove = "rm"
	// Env shows environment information used by lockbox
	Env = "var"
	// TOTPClip is the argument for copying totp codes to clipboard
	TOTPClip = Clip
	// TOTPMinimal is the argument for getting the short version of a code
	TOTPMinimal = "minimal"
	// TOTPList will list the totp-enabled entries
	TOTPList = List
	// TOTPOnce will perform like a normal totp request but not refresh
	TOTPOnce = "once"
	// CompletionsBash is the command to generate bash completions
	CompletionsBash = "bash"
	// Completions are used to generate shell completions
	Completions = "completions"
	// ReKey will rekey the underlying database
	ReKey = "rekey"
	// MultiLine handles multi-line inserts (when not piped)
	MultiLine = "multiline"
	// TOTPShow is for showing the TOTP token
	TOTPShow = Show
	// TOTPInsert is for inserting totp tokens
	TOTPInsert = Insert
	// JSON handles JSON outputs
	JSON = "json"
	// CompletionsZsh is the command to generate zsh completions
	CompletionsZsh = "zsh"
	// CompletionsFish is the command to generate fish completions
	CompletionsFish = "fish"
	// PasswordGenerate is the command to do password generation
	PasswordGenerate = "pwgen"
)

var (
	// CompletionTypes are shell completions that are known
	CompletionTypes = []string{CompletionsBash, CompletionsFish, CompletionsZsh}
	// ReKeyFlags are the flags used for re-keying
	ReKeyFlags = struct {
		KeyFile string
		NoKey   string
	}{"keyfile", "nokey"}
)

// ReKeyArgs is the base definition of re-keying args
type ReKeyArgs struct {
	NoKey   bool
	KeyFile string
}
