package clip_test

import (
	"testing"

	"github.com/seanenck/lockbox/internal/platform"
	"github.com/seanenck/lockbox/internal/platform/clip"
)

func TestNoClipboard(t *testing.T) {
	t.Setenv("LOCKBOX_CLIP_OSC52", "false")
	t.Setenv("LOCKBOX_CLIP_TIMEOUT", "")
	t.Setenv("LOCKBOX_CLIP_ENABLED", "false")
	_, err := clip.New()
	if err == nil || err.Error() != "clipboard is off" {
		t.Errorf("invalid error: %v", err)
	}
}

func TestMaxTime(t *testing.T) {
	t.Setenv("LOCKBOX_CLIP_ENABLED", "true")
	t.Setenv("LOCKBOX_CLIP_OSC52", "false")
	t.Setenv("LOCKBOX_PLATFORM", string(platform.Systems.LinuxWaylandSystem))
	t.Setenv("LOCKBOX_CLIP_TIMEOUT", "")
	c, err := clip.New()
	if err != nil {
		t.Errorf("invalid clipboard: %v", err)
	}
	if c.MaxTime != 45 {
		t.Error("invalid default")
	}
	t.Setenv("LOCKBOX_CLIP_TIMEOUT", "1")
	c, err = clip.New()
	if err != nil {
		t.Errorf("invalid clipboard: %v", err)
	}
	if c.MaxTime != 1 {
		t.Error("invalid default")
	}
	t.Setenv("LOCKBOX_CLIP_TIMEOUT", "-1")
	_, err = clip.New()
	if err == nil || err.Error() != "clipboard max time must be > 0" {
		t.Errorf("invalid max time error: %v", err)
	}
	t.Setenv("LOCKBOX_CLIP_TIMEOUT", "$&(+")
	_, err = clip.New()
	if err == nil || err.Error() != "strconv.Atoi: parsing \"$&(+\": invalid syntax" {
		t.Errorf("invalid max time error: %v", err)
	}
}

func TestClipboardInstances(t *testing.T) {
	t.Setenv("LOCKBOX_CLIP_ENABLED", "true")
	t.Setenv("LOCKBOX_CLIP_TIMEOUT", "")
	t.Setenv("LOCKBOX_CLIP_OSC52", "false")
	for _, item := range platform.Systems.List() {
		t.Setenv("LOCKBOX_PLATFORM", item)
		_, err := clip.New()
		if err != nil {
			t.Errorf("invalid clipboard: %v", err)
		}
	}
}

func TestOSC52(t *testing.T) {
	t.Setenv("LOCKBOX_CLIP_OSC52", "true")
	c, _ := clip.New()
	_, _, ok := c.Args(true)
	if ok {
		t.Error("invalid clipboard, should be an internal call")
	}
	_, _, ok = c.Args(false)
	if ok {
		t.Error("invalid clipboard, should be an internal call")
	}
}

func TestArgsOverride(t *testing.T) {
	t.Setenv("LOCKBOX_CLIP_PASTE_COMMAND", "abc xyz 111")
	t.Setenv("LOCKBOX_CLIP_OSC52", "false")
	t.Setenv("LOCKBOX_PLATFORM", string(platform.Systems.WindowsLinuxSystem))
	c, _ := clip.New()
	cmd, args, ok := c.Args(true)
	if cmd != "clip.exe" || len(args) != 0 || !ok {
		t.Error("invalid parse")
	}
	cmd, args, ok = c.Args(false)
	if cmd != "abc" || len(args) != 2 || args[0] != "xyz" || args[1] != "111" || !ok {
		t.Error("invalid parse")
	}
	t.Setenv("LOCKBOX_CLIP_COPY_COMMAND", "zzz lll 123")
	c, _ = clip.New()
	cmd, args, ok = c.Args(true)
	if cmd != "zzz" || len(args) != 2 || args[0] != "lll" || args[1] != "123" || !ok {
		t.Error("invalid parse")
	}
	cmd, args, ok = c.Args(false)
	if cmd != "abc" || len(args) != 2 || args[0] != "xyz" || args[1] != "111" || !ok {
		t.Error("invalid parse")
	}
	t.Setenv("LOCKBOX_CLIP_PASTE_COMMAND", "")
	t.Setenv("LOCKBOX_CLIP_COPY_COMMAND", "")
	c, _ = clip.New()
	cmd, args, ok = c.Args(true)
	if cmd != "clip.exe" || len(args) != 0 || !ok {
		t.Error("invalid parse")
	}
	cmd, args, ok = c.Args(false)
	if cmd != "powershell.exe" || len(args) != 2 || args[0] != "-command" || args[1] != "Get-Clipboard" || !ok {
		t.Errorf("invalid parse %s %v", cmd, args)
	}
}