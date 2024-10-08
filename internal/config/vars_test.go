package config_test

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/seanenck/lockbox/internal/config"
)

func checkYesNo(key string, t *testing.T, obj config.EnvironmentBool, onEmpty bool) {
	t.Setenv(key, "yes")
	c, err := obj.Get()
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if !c {
		t.Error("invalid setting")
	}
	t.Setenv(key, "")
	c, err = obj.Get()
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if c != onEmpty {
		t.Error("invalid setting")
	}
	t.Setenv(key, "no")
	c, err = obj.Get()
	if err != nil {
		t.Errorf("invalid error: %v", err)
	}
	if c {
		t.Error("invalid setting")
	}
	t.Setenv(key, "afoieae")
	_, err = obj.Get()
	if err == nil || err.Error() != fmt.Sprintf("invalid yes/no env value for %s", key) {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestColorSetting(t *testing.T) {
	checkYesNo("LOCKBOX_NOCOLOR", t, config.EnvNoColor, false)
}

func TestNoHook(t *testing.T) {
	checkYesNo("LOCKBOX_NOHOOKS", t, config.EnvNoHooks, false)
}

func TestInteractiveSetting(t *testing.T) {
	checkYesNo("LOCKBOX_INTERACTIVE", t, config.EnvInteractive, true)
}

func TestIsReadOnly(t *testing.T) {
	checkYesNo("LOCKBOX_READONLY", t, config.EnvReadOnly, false)
}

func TestIsOSC52(t *testing.T) {
	checkYesNo("LOCKBOX_CLIP_OSC52", t, config.EnvClipOSC52, false)
}

func TestIsNoTOTP(t *testing.T) {
	checkYesNo("LOCKBOX_NOTOTP", t, config.EnvNoTOTP, false)
}

func TestIsNoClip(t *testing.T) {
	checkYesNo("LOCKBOX_NOCLIP", t, config.EnvNoClip, false)
}

func TestIsNoGeneratePassword(t *testing.T) {
	checkYesNo("LOCKBOX_NOPWGEN", t, config.EnvNoPasswordGen, false)
}

func TestIsTitle(t *testing.T) {
	checkYesNo("LOCKBOX_PWGEN_TITLE", t, config.EnvPasswordGenTitle, true)
}

func TestDefaultCompletions(t *testing.T) {
	checkYesNo("LOCKBOX_DEFAULT_COMPLETION", t, config.EnvDefaultCompletion, false)
}

func TestTOTP(t *testing.T) {
	t.Setenv("LOCKBOX_TOTP", "abc")
	if config.EnvTOTPToken.Get() != "abc" {
		t.Error("invalid totp token field")
	}
	t.Setenv("LOCKBOX_TOTP", "")
	if config.EnvTOTPToken.Get() != "totp" {
		t.Error("invalid totp token field")
	}
}

func TestListVariables(t *testing.T) {
	known := make(map[string]struct{})
	for _, v := range config.ListEnvironmentVariables() {
		trim := strings.Split(strings.TrimSpace(v), " ")[0]
		if !strings.HasPrefix(trim, "LOCKBOX_") {
			t.Errorf("invalid env: %s", v)
		}
		if _, ok := known[trim]; ok {
			t.Errorf("invalid re-used env: %s", trim)
		}
		known[trim] = struct{}{}
	}
	l := len(known)
	if l != 34 {
		t.Errorf("invalid env count, outdated? %d", l)
	}
}

func TestReKey(t *testing.T) {
	if _, err := config.GetReKey([]string{"-nokey"}); err == nil || err.Error() != "a key or keyfile must be passed for rekey" {
		t.Errorf("failed: %v", err)
	}
	out, err := config.GetReKey([]string{})
	if err != nil {
		t.Errorf("failed: %v", err)
	}
	if out.NoKey || out.KeyFile != "" {
		t.Errorf("invalid args: %v", out)
	}
	out, err = config.GetReKey([]string{"-keyfile", "vars.go", "-nokey"})
	if err != nil {
		t.Errorf("failed: %v", err)
	}
	if !out.NoKey || out.KeyFile != "vars.go" {
		t.Errorf("invalid args: %v", out)
	}
}

func TestFormatTOTP(t *testing.T) {
	otp := config.EnvFormatTOTP.Get("otpauth://abc")
	if otp != "otpauth://abc" {
		t.Errorf("invalid totp token: %s", otp)
	}
	otp = config.EnvFormatTOTP.Get("abc")
	if otp != "otpauth://totp/lbissuer:lbaccount?algorithm=SHA1&digits=6&issuer=lbissuer&period=30&secret=abc" {
		t.Errorf("invalid totp token: %s", otp)
	}
	t.Setenv("LOCKBOX_TOTP_FORMAT", "test/%s")
	otp = config.EnvFormatTOTP.Get("abc")
	if otp != "test/abc" {
		t.Errorf("invalid totp token: %s", otp)
	}
	t.Setenv("LOCKBOX_TOTP_FORMAT", "")
	otp = config.EnvFormatTOTP.Get("abc")
	if otp != "otpauth://totp/lbissuer:lbaccount?algorithm=SHA1&digits=6&issuer=lbissuer&period=30&secret=abc" {
		t.Errorf("invalid totp token: %s", otp)
	}
}

func TestParseJSONMode(t *testing.T) {
	m, err := config.ParseJSONOutput()
	if m != config.JSONOutputs.Hash || err != nil {
		t.Error("invalid mode read")
	}
	t.Setenv("LOCKBOX_JSON_DATA", "hAsH ")
	m, err = config.ParseJSONOutput()
	if m != config.JSONOutputs.Hash || err != nil {
		t.Error("invalid mode read")
	}
	t.Setenv("LOCKBOX_JSON_DATA", "EMPTY")
	m, err = config.ParseJSONOutput()
	if m != config.JSONOutputs.Blank || err != nil {
		t.Error("invalid mode read")
	}
	t.Setenv("LOCKBOX_JSON_DATA", " PLAINtext ")
	m, err = config.ParseJSONOutput()
	if m != config.JSONOutputs.Raw || err != nil {
		t.Error("invalid mode read")
	}
	t.Setenv("LOCKBOX_JSON_DATA", "a")
	if _, err = config.ParseJSONOutput(); err == nil || err.Error() != "invalid JSON output mode: a" {
		t.Errorf("invalid error: %v", err)
	}
}

func TestClipboardMax(t *testing.T) {
	checkInt(config.EnvClipMax, "LOCKBOX_CLIP_MAX", "clipboard max time", 45, false, t)
}

func TestHashLength(t *testing.T) {
	checkInt(config.EnvHashLength, "LOCKBOX_JSON_DATA_HASH_LENGTH", "hash length", 0, true, t)
}

func TestMaxTOTP(t *testing.T) {
	checkInt(config.EnvMaxTOTP, "LOCKBOX_TOTP_MAX", "max totp time", 120, false, t)
}

func TestWordCount(t *testing.T) {
	checkInt(config.EnvPasswordGenCount, "LOCKBOX_PWGEN_COUNT", "word count", 8, false, t)
}

func checkInt(e config.EnvironmentInt, key, text string, def int, allowZero bool, t *testing.T) {
	t.Setenv(key, "")
	val, err := e.Get()
	if err != nil || val != def {
		t.Error("invalid read")
	}
	t.Setenv(key, "1")
	val, err = e.Get()
	if err != nil || val != 1 {
		t.Error("invalid read")
	}
	t.Setenv(key, "-1")
	zero := ""
	if allowZero {
		zero = "="
	}
	if _, err := e.Get(); err == nil || err.Error() != fmt.Sprintf("%s must be >%s 0", text, zero) {
		t.Errorf("invalid err: %v", err)
	}
	t.Setenv(key, "alk;ja")
	if _, err := e.Get(); err == nil || err.Error() != "strconv.Atoi: parsing \"alk;ja\": invalid syntax" {
		t.Errorf("invalid err: %v", err)
	}
	t.Setenv(key, "0")
	if allowZero {
		val, err = e.Get()
		if err != nil || val != 0 {
			t.Error("invalid read")
		}
	} else {
		if _, err := e.Get(); err == nil || err.Error() != fmt.Sprintf("%s must be > 0", text) {
			t.Errorf("invalid err: %v", err)
		}
	}
}

func TestEnvironDefinitions(t *testing.T) {
	os.Clearenv()
	vals := config.ListEnvironmentVariables()
	expect := make(map[string]struct{})
	found := false
	for _, val := range vals {
		found = true
		env := strings.Split(strings.TrimSpace(val), "\n")[0]
		if !strings.HasPrefix(env, "LOCKBOX_") {
			t.Errorf("invalid env var: %s", env)
		}
		if env == "LOCKBOX_ENV" {
			continue
		}
		t.Setenv(env, "test")
		expect[env] = struct{}{}
	}
	if !found {
		t.Errorf("no environment variables found?")
	}
	read := config.Environ()
	if len(read) != len(expect) {
		t.Errorf("invalid environment variable info: %d != %d", len(expect), len(read))
	}
	for k := range expect {
		found := false
		for _, r := range read {
			if r == fmt.Sprintf("%s=test", k) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("unable to find env: %s", k)
		}
	}
}

func TestCanColor(t *testing.T) {
	os.Clearenv()
	if can, _ := config.CanColor(); !can {
		t.Error("should be able to color")
	}
	for raw, expect := range map[string]bool{
		"INTERACTIVE": true,
		"NOCOLOR":     false,
	} {
		os.Clearenv()
		key := fmt.Sprintf("LOCKBOX_%s", raw)
		t.Setenv(key, "yes")
		if can, _ := config.CanColor(); can != expect {
			t.Errorf("expect != actual: %s", key)
		}
		t.Setenv(key, "no")
		if can, _ := config.CanColor(); can == expect {
			t.Errorf("expect == actual: %s", key)
		}
	}
	os.Clearenv()
	t.Setenv("NO_COLOR", "1")
	if can, _ := config.CanColor(); can {
		t.Error("should NOT be able to color")
	}
}
