package platform_test

import (
	"fmt"
	"testing"

	"github.com/seanenck/lockbox/internal/config"
	"github.com/seanenck/lockbox/internal/platform"
)

func TestNoClipboard(t *testing.T) {
	t.Setenv("LOCKBOX_CLIP_OSC52", "no")
	t.Setenv("LOCKBOX_CLIP_MAX", "")
	t.Setenv("LOCKBOX_NOCLIP", "yes")
	_, err := platform.NewClipboard()
	if err == nil || err.Error() != "clipboard is off" {
		t.Errorf("invalid error: %v", err)
	}
}

func TestMaxTime(t *testing.T) {
	t.Setenv("LOCKBOX_NOCLIP", "no")
	t.Setenv("LOCKBOX_CLIP_OSC52", "no")
	t.Setenv("LOCKBOX_PLATFORM", string(config.Platforms.LinuxWaylandPlatform))
	t.Setenv("LOCKBOX_CLIP_MAX", "")
	c, err := platform.NewClipboard()
	if err != nil {
		t.Errorf("invalid clipboard: %v", err)
	}
	if c.MaxTime != 45 {
		t.Error("invalid default")
	}
	t.Setenv("LOCKBOX_CLIP_MAX", "1")
	c, err = platform.NewClipboard()
	if err != nil {
		t.Errorf("invalid clipboard: %v", err)
	}
	if c.MaxTime != 1 {
		t.Error("invalid default")
	}
	t.Setenv("LOCKBOX_CLIP_MAX", "-1")
	_, err = platform.NewClipboard()
	if err == nil || err.Error() != "clipboard max time must be > 0" {
		t.Errorf("invalid max time error: %v", err)
	}
	t.Setenv("LOCKBOX_CLIP_MAX", "$&(+")
	_, err = platform.NewClipboard()
	if err == nil || err.Error() != "strconv.Atoi: parsing \"$&(+\": invalid syntax" {
		t.Errorf("invalid max time error: %v", err)
	}
}

func TestClipboardInstances(t *testing.T) {
	t.Setenv("LOCKBOX_NOCLIP", "no")
	t.Setenv("LOCKBOX_CLIP_MAX", "")
	t.Setenv("LOCKBOX_CLIP_OSC52", "no")
	for _, item := range config.Platforms.List() {
		t.Setenv("LOCKBOX_PLATFORM", item)
		_, err := platform.NewClipboard()
		if err != nil {
			t.Errorf("invalid clipboard: %v", err)
		}
	}
}

func TestOSC52(t *testing.T) {
	t.Setenv("LOCKBOX_CLIP_OSC52", "yes")
	c, _ := platform.NewClipboard()
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
	t.Setenv("LOCKBOX_CLIP_PASTE", "abc xyz 111")
	t.Setenv("LOCKBOX_CLIP_OSC52", "no")
	t.Setenv("LOCKBOX_PLATFORM", string(config.Platforms.WindowsLinuxPlatform))
	c, _ := platform.NewClipboard()
	cmd, args, ok := c.Args(true)
	if cmd != "clip.exe" || len(args) != 0 || !ok {
		t.Error("invalid parse")
	}
	cmd, args, ok = c.Args(false)
	if cmd != "abc" || len(args) != 2 || args[0] != "xyz" || args[1] != "111" || !ok {
		t.Error("invalid parse")
	}
	t.Setenv("LOCKBOX_CLIP_COPY", "zzz lll 123")
	c, _ = platform.NewClipboard()
	cmd, args, ok = c.Args(true)
	if cmd != "zzz" || len(args) != 2 || args[0] != "lll" || args[1] != "123" || !ok {
		t.Error("invalid parse")
	}
	cmd, args, ok = c.Args(false)
	if cmd != "abc" || len(args) != 2 || args[0] != "xyz" || args[1] != "111" || !ok {
		t.Error("invalid parse")
	}
	t.Setenv("LOCKBOX_CLIP_PASTE", "")
	t.Setenv("LOCKBOX_CLIP_COPY", "")
	c, _ = platform.NewClipboard()
	cmd, args, ok = c.Args(true)
	if cmd != "clip.exe" || len(args) != 0 || !ok {
		t.Error("invalid parse")
	}
	cmd, args, ok = c.Args(false)
	if cmd != "powershell.exe" || len(args) != 2 || args[0] != "-command" || args[1] != "Get-Clipboard" || !ok {
		fmt.Println(args)
		t.Error("invalid parse")
	}
}
