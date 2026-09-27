package systeminfo

import (
	"context"
	"errors"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"unicode"
	"unicode/utf16"
)

var serviceNamePattern = regexp.MustCompile(`^[A-Za-z0-9_.@:-]{1,128}$`)

func ValidServiceName(name string) bool {
	if name == "" || len(utf16.Encode([]rune(name))) > 256 || strings.ContainsAny(name, `/\`) {
		return false
	}
	for _, r := range name {
		if unicode.IsControl(r) {
			return false
		}
	}
	return true
}

func ServiceAction(ctx context.Context, name, action string) error {
	if !ValidServiceName(name) {
		return errors.New("invalid service name")
	}
	if action != "start" && action != "stop" && action != "restart" {
		return errors.New("unsupported service action")
	}
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux":
		if !serviceNamePattern.MatchString(name) || !strings.HasSuffix(name, ".service") {
			return errors.New("service action must identify a service unit")
		}
		cmd = exec.CommandContext(ctx, "systemctl", action, "--", name)
	case "windows":
		return windowsServiceAction(ctx, name, action)
	case "darwin":
		return errors.New("launchd service changes require an explicit plist target")
	default:
		return errors.New("service backend unavailable")
	}
	return cmd.Run()
}
